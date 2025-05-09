package trivia

import (
	"app/domain/tts"
	"app/env"
	"app/messages"
	"bytes"
	"context"
	_ "embed"
	"errors"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"lib/aws"
	discord2 "lib/discord"
	"lib/logging"
	util2 "lib/util"
	"lib/util/arrayutil"
	"strings"
	"time"
)

type Trivia struct {
	bot             *discord2.Bot
	channelID       string
	onDisposed      func()
	vm              *discord2.VoiceManager
	tts             *tts.Client
	logger          *zap.Logger
	state           *State
	AnswerReceived  chan string
	PlayerNominated chan *discordgo.User
	isStarted       bool
	s3              *aws.S3
}

//go:embed static/intro.mp3
var intro []byte

//go:embed static/good.mp3
var good []byte

//go:embed static/wrong.mp3
var wrong []byte

const questionTimeout = time.Second * 30

func New(s3 *aws.S3, bot *discord2.Bot, tts *tts.Client, channelID string, onDisposed func()) (*Trivia, error) {
	vm, err := discord2.NewManager(bot, channelID, onDisposed)
	if err != nil {
		return nil, err
	}

	var players []*discordgo.User

	members, err := bot.ListVoiceChannelMembers(env.Env.GuildId, channelID)
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
		bot:             bot,
		channelID:       channelID,
		onDisposed:      onDisposed,
		vm:              vm,
		tts:             tts,
		logger:          logging.Get().Named("trivia").With(zap.String("channelID", channelID)),
		state:           NewState(s3, players),
		AnswerReceived:  make(chan string),
		PlayerNominated: make(chan *discordgo.User),
		s3:              s3,
	}

	return trivia, nil
}

func (t *Trivia) Start() error {
	if t.isStarted {
		return nil
	}

	msg := arrayutil.RandomElement(messages.Messages.Trivia.Start)
	welcomeSentence := util2.ApplyTokens(msg, map[string]string{
		"MEMBERS_COUNT": util2.PlayerCountSentence(len(t.state.players)),
	})

	err := t.playIntro()
	if err != nil {
		t.logger.Error("failed to play intro", zap.Error(err))
	}

	err = t.speak(welcomeSentence)
	if err != nil {
		t.logger.Error("failed to speak welcome sentence", zap.Error(err))
	}

	startingPlayer := arrayutil.RandomElement(t.state.players)
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

	switch question.Type {
	case TrueFalse:
		validAnswers = messages.Messages.Trivia.ValidAnswer.Boolean

	case MultipleChoice:
		validAnswers = messages.Messages.Trivia.ValidAnswer.Multiple

	default:
		t.logger.Error("invalid question type", zap.String("type", string(question.Type)))
		return
	}

	q := question.ForSpeaking()
	options := strings.Join(arrayutil.Shuffle(question.Options()), ", ")

	var name string
	friend, ok := discord2.Friends[t.state.currentPlayer.ID]
	if ok {
		name = friend.Nickname
	} else {
		name = t.state.currentPlayer.GlobalName
	}

	tokens := map[string]string{
		"ANSWER":   question.Correct,
		"NAME":     name,
		"QUESTION": q,
		"OPTIONS":  options,
		"MENTION":  t.state.currentPlayer.Mention(),
	}

	incorrectAnswers := question.IncorrectAnswerMessages

	if !arrayutil.IsValidArray(validAnswers) || !arrayutil.IsValidArray(incorrectAnswers) {
		t.NextQuestion()
		return
	}

	validAnswerPhrase := util2.ApplyTokens(arrayutil.RandomElement(validAnswers), tokens)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(questionTimeout))
	defer cancel()

	err = t.speakQuestion()
	if err != nil {
		t.logger.Error("failed to speak question", zap.Error(err))
		return
	}

	messageContent := util2.ApplyTokens(arrayutil.RandomElement(messages.Messages.Trivia.QuestionMessages), tokens)
	m, err := t.bot.ChannelMessageSendComplex(t.channelID, &discordgo.MessageSend{
		Components: GetQuestionComponent(question, nil),
		Content:    messageContent,
	})
	if err != nil {
		t.logger.Error("failed to send question", zap.Error(err))
		t.Dispose()
		return
	}
	defer func() {
		err = t.bot.ChannelMessageDelete(t.channelID, m.ID)
		if err != nil {
			t.logger.Error("failed to delete question message", zap.Error(err))
		}
	}()

	select {
	case answer := <-t.AnswerReceived:
		t.logger.Info("answer received", zap.String("answer", answer))

		component := GetQuestionComponent(question, &QuestionComponentOpts{
			SelectedAnswer: answer,
		})
		_, err = t.bot.ChannelMessageEditComplex(&discordgo.MessageEdit{
			ID:         m.ID,
			Channel:    t.channelID,
			Components: &component,
			Content:    &messageContent,
		})
		if err != nil {
			t.logger.Error("failed to edit question", zap.Error(err))
		}

		if answer == question.Correct {
			t.handleCorrectAnswer(validAnswerPhrase)
		} else {
			t.handleIncorrectAnswer()
		}

		return

	case <-ctx.Done():
		t.logger.Info("answer timeout")
		t.handleQuestionTimeout()

		return

	case <-t.vm.Disposed:
		t.logger.Info("voice connection disposed")

		return
	}

}

func (t *Trivia) maybeSayFunFact() {
	q := t.state.currentQuestion

	if q != nil && util2.RandomBool() && len(q.FunFacts) > 0 {
		err := t.speak(arrayutil.RandomElement(q.FunFacts))
		if err != nil {
			t.logger.Error("failed to speak fun fact", zap.Error(err))
		}
	}
}

// TODO Implement
func (t *Trivia) handleQuestionTimeout() {

}

// TODO Send message to channel with points
func (t *Trivia) finish() {
	if !t.isStarted {
		return
	}

	winners := t.state.GetWinners()

	var winnerMsg string

	if len(winners) == 0 {
		winnerMsg = arrayutil.RandomElement(messages.Messages.Trivia.NoMoreQuestionsNoWinner)
	} else if len(winners) > 1 {
		winnerMsg = arrayutil.RandomElement(messages.Messages.Trivia.NoMoreQuestionsWinner)
	} else {
		winnerMsg = arrayutil.RandomElement(messages.Messages.Trivia.NoMoreQuestionsDraw)
	}

	winnerMsg = util2.ApplyTokens(winnerMsg, map[string]string{
		"MENTION": strings.Join(arrayutil.Map(winners, func(w *discordgo.User) string {
			return w.Mention()
		}), ", "),
	})

	err := t.speak(winnerMsg)
	if err != nil {
		t.logger.Error("failed to speak winner", zap.Error(err))
	}

	t.SendPointsMessage()

	t.isStarted = false
	t.Dispose()
}

func (t *Trivia) SendPointsMessage() {
	t.bot.SendMessageAndForget(t.channelID, t.state.GetPointsMessageContents())
}

func (t *Trivia) handleIncorrectAnswer() {
	question := t.state.currentQuestion
	if question == nil {
		return
	}

	member := t.state.GetCurrentPlayerMemberDefinition()
	if member == nil {
		return
	}

	phrases := question.GetInvalidAnswerPhraseParts(member.Nickname)

	t.logger.Info("handling incorrect answer", zap.Strings("phrases", phrases))
	err := t.playBadSound()
	if err != nil {
		t.logger.Error("failed to play bad sound", zap.Error(err))
	}

	t.logger.Info("incorrect answer")

	err = t.speakMultiple(phrases...)
	if err != nil {
		t.logger.Error("failed to speak incorrect answer", zap.Error(err), zap.Strings("phrases", phrases))
	}

	t.maybeSayFunFact()
	t.state.ChangePlayerToRandom()
	go t.NextQuestion()
}

func (t *Trivia) handleCorrectAnswer(phrase string) {
	t.logger.Info("handling correct answer", zap.String("answer", phrase))
	err := t.playGoodSound()
	if err != nil {
		t.logger.Error("failed to play good sound", zap.Error(err))
	}

	t.state.AddPointToCurrentPlayer()

	t.logger.Info("correct answer")
	err = t.speak(phrase)
	if err != nil {
		t.logger.Error("failed to speak correct answer", zap.String("phrase", phrase), zap.Error(err))
	}

	t.maybeSayFunFact()

	if t.state.HasMoreQuestions() {
		t.nominateForNextQuestion()
	} else {
		t.finish()
	}
}

func (t *Trivia) nominateForNextQuestion() {
	m, err := t.bot.ChannelMessageSendComplex(t.channelID, &discordgo.MessageSend{
		Components: GetQuestionNominationComponent(),
		Content: util2.ApplyTokens(messages.Messages.Trivia.PickNextPlayer, map[string]string{
			"MENTION": t.state.currentPlayer.Mention(),
		}),
	})
	if err != nil {
		t.Dispose()
		return
	}
	defer func() {
		err = t.bot.ChannelMessageDelete(t.channelID, m.ID)
		if err != nil {
			t.logger.Error("failed to delete nomination message", zap.Error(err))
		}
	}()

	select {
	case player := <-t.PlayerNominated:
		t.state.ChangePlayer(player)
		go t.NextQuestion()

		return

	case <-time.After(time.Minute * 1):
		t.logger.Info("nomination timeout")
		t.state.ChangePlayerToRandom()
		go t.NextQuestion()

		return

	case <-t.vm.Disposed:
		return
	}

}

func (t *Trivia) playIntro() error {
	return t.speakNonDcaBytes(intro)
}

func (t *Trivia) speak(text string) error {
	v, err := GetVoice(t.s3, text)

	if err != nil {
		return err
	}

	speaker := discord2.NewDcaSpeaker(v)

	return t.vm.Speak(speaker)
}

func (t *Trivia) speakMultiple(texts ...string) error {
	var voices []*discord2.DcaSpeaker

	for _, text := range texts {
		v, err := GetVoice(t.s3, text)
		if err != nil {
			return err
		}
		voices = append(voices, discord2.NewDcaSpeaker(v))
	}

	for _, speaker := range voices {
		err := t.vm.Speak(speaker)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Trivia) speakQuestion() error {
	friend, ok := discord2.Friends[t.state.currentPlayer.ID]
	if !ok {
		return errors.New("friend not found")
	}

	tokens := map[string]string{
		"MENTION": t.state.currentPlayer.Mention(),
		"NAME":    friend.Nickname,
	}

	var questionPhraseTemplate string
	if t.state.previousPlayer != nil && t.state.currentPlayer.ID == t.state.previousPlayer.ID {
		t.logger.Info("current player is previous player")
		questionPhraseTemplate = arrayutil.RandomElement(messages.Messages.Trivia.CurrentPlayerNextQuestion)
	} else {
		t.logger.Info("current player is not previous player")
		questionPhraseTemplate = arrayutil.RandomElement(messages.Messages.Trivia.NextPlayerQuestion)
	}
	questionPhrase := util2.ApplyTokens(questionPhraseTemplate, tokens)

	return t.speakMultiple(questionPhrase, t.state.currentQuestion.ForSpeaking())
}

func (t *Trivia) playGoodSound() error {
	return t.speakNonDcaBytes(good)
}

func (t *Trivia) playBadSound() error {
	return t.speakNonDcaBytes(wrong)
}

func (t *Trivia) speakNonDcaBytes(v []byte) error {
	speaker := discord2.NewNonDcaSpeaker(bytes.NewReader(v))

	return t.vm.Speak(speaker)
}

func (t *Trivia) Dispose() {
	t.vm.Dispose()
}

func (t *Trivia) HandleBoolean(answer string) error {
	select {
	case t.AnswerReceived <- answer:
		return nil

	default:
		return errors.New("answer channel is not handling answer")
	}
}

func (t *Trivia) HandleChoice(answerIndex int) error {
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
