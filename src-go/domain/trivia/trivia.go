package trivia

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"github.com/bwmarrin/discordgo"
	dca2 "github.com/jonas747/dca/v2"
	"go.uber.org/zap"
	"io"
	"src-go/discord"
	"src-go/domain/tts"
	"src-go/domain/voice"
	"src-go/logging"
	"src-go/messages"
	"src-go/util"
	"strings"
	"sync"
	"time"
)

type Trivia struct {
	session         *discordgo.Session
	channelID       string
	onDisposed      func()
	vm              *voice.Manager
	tts             *tts.Client
	logger          *zap.Logger
	speakerLock     sync.Mutex
	state           *State
	AnswerReceived  chan string
	PlayerNominated chan *discordgo.User
	isStarted       bool
	ctx             context.Context
}

//go:embed static/intro.mp3
var intro []byte

//go:embed static/good.mp3
var good []byte

//go:embed static/wrong.mp3
var wrong []byte

const questionTimeout = time.Second * 30

func New(ctx context.Context, session *discordgo.Session, tts *tts.Client, channelID string, onDisposed func()) (*Trivia, error) {
	vm, err := voice.NewManager(session, channelID, onDisposed)
	if err != nil {
		return nil, err
	}

	var players []*discordgo.User

	members, err := discord.ListVoiceChannelMembers(session, channelID)
	if err != nil {
		return nil, err
	}
	for _, member := range members {
		if member.User.Bot {
			continue
		}

		players = append(players, member.User)
	}

	trivia := &Trivia{
		session:         session,
		channelID:       channelID,
		onDisposed:      onDisposed,
		vm:              vm,
		tts:             tts,
		logger:          logging.Get().Named("trivia").With(zap.String("channelID", channelID)),
		state:           NewState(ctx, players),
		AnswerReceived:  make(chan string),
		PlayerNominated: make(chan *discordgo.User),
		ctx:             ctx,
	}

	return trivia, nil
}

func (t *Trivia) Start() error {
	if t.isStarted {
		return nil
	}

	msg := util.RandomElement(messages.Messages.Trivia.Start)
	welcomeSentence := util.ApplyTokens(msg, map[string]string{
		"MEMBERS_COUNT": util.PlayerCountSentence(len(t.state.players)),
	})

	err := t.playIntro()
	if err != nil {
		t.logger.Error("failed to play intro", zap.Error(err))
	}

	err = t.speak(welcomeSentence)
	if err != nil {
		t.logger.Error("failed to speak welcome sentence", zap.Error(err))
	}

	startingPlayer := util.RandomElement(t.state.players)
	t.state.SetStartingPlayer(startingPlayer)

	go t.NextQuestion()

	t.isStarted = true

	return nil
}

func (t *Trivia) NextQuestion() {
	question, err := t.state.GetNextQuestion()
	if err != nil {
		if errors.Is(err, ErrNoMoreQuestions) {
			t.finish()
			return
		}

		t.logger.Error("failed to get next question", zap.Error(err))
		return
	}
	var validAnswers []string
	var invalidAnwsers []string

	switch question.Type {
	case TrueFalse:
		validAnswers = messages.Messages.Trivia.ValidAnswer.Boolean
		invalidAnwsers = messages.Messages.Trivia.InvalidAnswer.Boolean

	case MultipleChoice:
		validAnswers = messages.Messages.Trivia.ValidAnswer.Multiple
		invalidAnwsers = messages.Messages.Trivia.InvalidAnswer.Multiple

	default:
		t.logger.Error("invalid question type", zap.String("type", string(question.Type)))
		return
	}

	q := question.QuestionForSpeaking()
	options := strings.Join(util.Shuffle(question.Options()), ", ")

	tokens := map[string]string{
		"ANSWER":   question.Correct,
		"NAME":     t.state.currentPlayer.GlobalName,
		"QUESTION": q,
		"OPTIONS":  options,
		"MENTION":  t.state.currentPlayer.Mention(),
	}
	validAnswerPhrase := util.ApplyTokens(util.RandomElement(validAnswers), tokens)
	invalidPhraseAnswer := util.ApplyTokens(util.RandomElement(invalidAnwsers), tokens)

	var questionPhraseTemplate string
	if t.state.previousPlayer != nil && t.state.currentPlayer.ID == t.state.previousPlayer.ID {
		t.logger.Info("current player is previous player")
		questionPhraseTemplate = util.RandomElement(messages.Messages.Trivia.CurrentPlayerNextQuestion)
	} else {
		t.logger.Info("current player is not previous player")
		questionPhraseTemplate = util.RandomElement(messages.Messages.Trivia.NextPlayerQuestion)
	}
	questionPhrase := util.ApplyTokens(questionPhraseTemplate, tokens)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(questionTimeout))
	defer cancel()

	err = t.speak(questionPhrase)
	if err != nil {
		t.logger.Error("failed to speak question", zap.Error(err))
		return
	}

	m, err := t.session.ChannelMessageSendComplex(t.channelID, &discordgo.MessageSend{
		Components: GetQuestionComponent(question),
		Content:    util.ApplyTokens(util.RandomElement(messages.Messages.Trivia.QuestionMessages), tokens),
	})
	if err != nil {
		t.logger.Error("failed to send question", zap.Error(err))
		t.Dispose()
		return
	}
	defer func() {
		err = t.session.ChannelMessageDelete(t.channelID, m.ID)
		if err != nil {
			t.logger.Error("failed to delete question message", zap.Error(err))
		}
	}()

	select {
	case answer := <-t.AnswerReceived:
		t.logger.Info("answer received", zap.String("answer", answer))

		if answer == question.Correct {
			go t.handleCorrectAnswer(validAnswerPhrase)
		} else {
			go t.handleIncorrectAnswer(invalidPhraseAnswer)
		}

	case <-ctx.Done():
		t.logger.Info("answer timeout")
		go t.handleQuestionTimeout()

	case <-t.vm.Disposed:
		t.logger.Info("voice connection disposed")
	}

}

// TODO Implement
func (t *Trivia) handleQuestionTimeout() {

}

func (t *Trivia) finish() {
	if !t.isStarted {
		return
	}

	winners := t.state.GetWinners()

	var winnerMsg string

	if len(winners) == 0 {
		winnerMsg = util.RandomElement(messages.Messages.Trivia.NoMoreQuestionsNoWinner)
	} else if len(winners) > 1 {
		winnerMsg = util.RandomElement(messages.Messages.Trivia.NoMoreQuestionsWinner)
	} else {
		winnerMsg = util.RandomElement(messages.Messages.Trivia.NoMoreQuestionsDraw)
	}

	winnerMsg = util.ApplyTokens(winnerMsg, map[string]string{
		"MENTION": strings.Join(util.Map(winners, func(w *discordgo.User) string {
			return w.Mention()
		}), ", "),
	})

	err := t.speak(winnerMsg)
	if err != nil {
		t.logger.Error("failed to speak winner", zap.Error(err))
	}

	t.isStarted = false
	t.Dispose()
}

func (t *Trivia) handleIncorrectAnswer(invalidPhraseAnswer string) {
	err := t.playBadSound()
	if err != nil {
		t.logger.Error("failed to play bad sound", zap.Error(err))
	}

	t.logger.Info("incorrect answer")
	err = t.speak(invalidPhraseAnswer)
	if err != nil {
		t.logger.Error("failed to speak incorrect answer", zap.Error(err))
	}

	t.state.ChangePlayerToRandom()
	t.NextQuestion()
}

func (t *Trivia) handleCorrectAnswer(validAnswerPhrase string) {
	err := t.playGoodSound()
	if err != nil {
		t.logger.Error("failed to play good sound", zap.Error(err))
	}

	t.state.AddPointToCurrentPlayer()

	t.logger.Info("correct answer")
	err = t.speak(validAnswerPhrase)
	if err != nil {
		t.logger.Error("failed to speak correct answer", zap.Error(err))
	}

	t.nominateForNextQuestion()

}

func (t *Trivia) nominateForNextQuestion() {
	m, err := t.session.ChannelMessageSendComplex(t.channelID, &discordgo.MessageSend{
		Components: GetQuestionNominationComponent(),
		Content: util.ApplyTokens(messages.Messages.Trivia.PickNextPlayer, map[string]string{
			"MENTION": t.state.currentPlayer.Mention(),
		}),
	})
	if err != nil {
		t.Dispose()
		return
	}
	defer func() {
		err = t.session.ChannelMessageDelete(t.channelID, m.ID)
		if err != nil {
			t.logger.Error("failed to delete nomination message", zap.Error(err))
		}
	}()

	select {
	case player := <-t.PlayerNominated:
		t.state.ChangePlayer(player)
		go t.NextQuestion()

	case <-time.After(time.Minute * 1):
		t.logger.Info("nomination timeout")
		t.state.ChangePlayerToRandom()
		go t.NextQuestion()

	case <-t.vm.Disposed:
		return
	}

}

func (t *Trivia) playIntro() error {
	return t.speakNonDcaBytes(intro)
}

func (t *Trivia) speak(text string) error {
	v, err := GetVoice(t.ctx, text)

	if err != nil {
		return err
	}

	return t.speakDca(v)
}

func (t *Trivia) playGoodSound() error {
	return t.speakNonDcaBytes(good)
}

func (t *Trivia) playBadSound() error {
	return t.speakNonDcaBytes(wrong)
}

func (t *Trivia) speakDca(voice io.ReadCloser) error {
	t.speakerLock.Lock()
	defer t.speakerLock.Unlock()

	vc, err := t.vm.VoiceConnection()
	if err != nil {
		return err
	}

	decoder := dca2.NewDecoder(voice)

	err = vc.Speaking(true)
	defer vc.Speaking(false)
	if err != nil {
		return err
	}

	for {
		frame, err := decoder.OpusFrame()

		if err != nil {
			if err == io.EOF {
				return nil
			}

			return err
		}

		select {
		case vc.OpusSend <- frame:
			continue
		case <-t.vm.Disposed:
			return nil
		}
	}
}

func (t *Trivia) speakNonDcaBytes(voice []byte) error {
	t.speakerLock.Lock()
	defer t.speakerLock.Unlock()

	vc, err := t.vm.VoiceConnection()
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(voice)
	stream, err := dca2.EncodeMem(buf, dca2.StdEncodeOptions)
	if err != nil {
		return err
	}
	defer stream.Cleanup()

	err = vc.Speaking(true)
	defer vc.Speaking(false)
	if err != nil {
		return err
	}

	for {
		frame, err := stream.OpusFrame()
		if err != nil {
			break
		}

		select {
		case vc.OpusSend <- frame:
			continue
		case <-t.vm.Disposed:
			return nil
		}
	}

	return nil
}

func (t *Trivia) Dispose() {
	t.vm.Dispose()
}

func (t *Trivia) HandleAnswer(answerIndex int) error {
	options := t.state.currentQuestion.Options()

	if answerIndex > len(options) {
		return errors.New("invalid answer index")
	}

	answer := options[answerIndex]

	select {
	case t.AnswerReceived <- answer:
		return nil

	default:
		return errors.New("answer channel is not handling answer")
	}
}

func (t *Trivia) GetState() *State {
	return t.state
}
