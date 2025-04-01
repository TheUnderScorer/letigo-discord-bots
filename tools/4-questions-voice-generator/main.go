package main

import (
	"app/dca"
	"app/discord"
	"app/domain/trivia"
	tts2 "app/domain/tts"
	"app/env"
	"app/messages"
	util2 "app/util"
	"app/util/arrayutil"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/avast/retry-go/v4"
	"github.com/charmbracelet/log"
	"github.com/joho/godotenv"
	"io"
	"os"
	"sync"
	"time"
	"tools/shared/util"
)

var resultPath = "4-questions-voice-generator/result"

func main() {
	messages.Init()
	err := godotenv.Load("../.env")
	if err != nil {
		log.Warn("error loading .env file", err)
	}

	env.Init()

	client := tts2.NewClient()

	var wg sync.WaitGroup
	limit := make(chan bool, 5)

	questions := util.ReadStepResult[trivia.Question]("3-questions-enhancer")

	wg.Add(len(questions))

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Hour*3))
	defer cancel()

	err = os.MkdirAll(resultPath, 0777)
	if err != nil {
		if os.IsExist(err) {
			log.Info("result directory already exists")
		} else {
			log.Fatal("failed to create result directory", err)
		}
	}

	length := len(questions)
	for i, q := range questions {
		go func() {
			limit <- true
			defer func() {
				wg.Done()
				<-limit
			}()

			pos := i + 1
			log.Infof("Generating %d/%d", pos, length)

			generateForQuestion(ctx, q, client)

			log.Infof("Generated %d/%d", pos, length)
		}()
	}

	wg.Add(1)

	go func() {
		defer wg.Done()
		generateMiscPhrases(ctx, client)
	}()

	wg.Wait()
}

func generateForQuestion(ctx context.Context, question *trivia.Question, client *tts2.Client) {
	var sentences []string

	if question.Type == trivia.MultipleChoice {
		sentences = append(sentences, question.IncorrectAnswerMessages...)
	}

	sentences = append(sentences, question.ForSpeaking())

	if arrayutil.IsValidArray(question.FunFacts) {
		sentences = append(sentences, question.FunFacts...)
	}

	generateSentences(ctx, sentences, client)
}

func generateMiscPhrases(ctx context.Context, client *tts2.Client) {
	var phrases []string

	phrases = append(phrases, messages.Messages.Trivia.NoMoreQuestionsDraw...)
	phrases = append(phrases, messages.Messages.Trivia.NoMoreQuestionsNoWinner...)

	for _, m := range messages.Messages.Trivia.Start {
		for i := range 12 {
			phrases = append(phrases, util2.ApplyTokens(m, map[string]string{
				"MEMBERS_COUNT": util2.PlayerCountSentence(i),
			}))
		}
	}

	for _, friend := range discord.Friends {
		var todoPhrases []string
		todoPhrases = append(todoPhrases, messages.Messages.Trivia.NoMoreQuestionsWinner...)
		todoPhrases = append(todoPhrases, messages.Messages.Trivia.InvalidAnswer.MultipleLeft...)
		todoPhrases = append(todoPhrases, messages.Messages.Trivia.InvalidAnswer.BooleanTrue...)
		todoPhrases = append(todoPhrases, messages.Messages.Trivia.InvalidAnswer.BooleanFalse...)
		todoPhrases = append(todoPhrases, messages.Messages.Trivia.ValidAnswer.Boolean...)
		todoPhrases = append(todoPhrases, messages.Messages.Trivia.ValidAnswer.Multiple...)
		todoPhrases = append(todoPhrases, messages.Messages.Trivia.NextPlayerQuestion...)
		todoPhrases = append(todoPhrases, messages.Messages.Trivia.CurrentPlayerNextQuestion...)

		tokens := map[string]string{
			"MENTION": friend.Nickname,
			"NAME":    friend.Nickname,
		}

		for _, phrase := range todoPhrases {
			phrases = append(phrases, util2.ApplyTokens(phrase, tokens))
		}

		phrases = append(phrases, util2.ApplyTokens(messages.Messages.Trivia.PickNextPlayer, tokens))
	}

	generateSentences(ctx, phrases, client)
}

func generateSentences(ctx context.Context, sentences []string, client *tts2.Client) {
	var wg sync.WaitGroup
	limit := make(chan bool, 3)

	for _, sentence := range sentences {
		wg.Add(1)
		go func() {
			limit <- true
			defer func() {
				<-limit
				wg.Done()
			}()
			err := generateSentence(ctx, sentence, resultPath, client)
			if err != nil {
				log.Errorf("Failed to generateForQuestion sentence: %s", err)
			}
		}()
	}

	wg.Wait()
}

// generateSentence generates a sentence as a .dca file
func generateSentence(ctx context.Context, sentence string, dir string, client *tts2.Client) error {
	const sentenceCount = 4
	const maxAttempts = 5

	i := 0

	logger := log.WithPrefix(sentence)

	hash := util2.Hash(sentence)
	sentenceDirPath := fmt.Sprintf("%s/%s", dir, hash)
	err := os.Mkdir(sentenceDirPath, 0777)
	if err != nil {
		if os.IsExist(err) {
			dirContents, err := os.ReadDir(dir)
			if err != nil {
				return err
			}

			// 4 .dca. files and .meta file
			if len(dirContents) == 5 {
				log.Infof("sentence already exists with all required files: %s", sentence)

				return nil
			}

			log.Warnf("sentence already exists: %s", sentence)
		} else {
			return err
		}
	}

	onRetry := retry.OnRetry(func(attempt uint, err error) {
		logger.Warnf("Attempt %d to generateForQuestion voice failed: %v", attempt, err)
	})

	for i < sentenceCount {
		fileName := fmt.Sprintf("%s/%d.dca", sentenceDirPath, i)

		if _, err := os.Stat(fileName); err == nil {
			i++
			logger.Warnf("File already exists: %s", fileName)
			continue
		}

		voice, err := retry.DoWithData(func() ([]byte, error) {
			return client.TextToVoice(ctx, &tts2.TextToVoiceRequest{
				Text:    sentence,
				Speaker: tts2.SpeakerTadeusz,
			})
		}, retry.Context(ctx), retry.Attempts(maxAttempts), retry.Delay(time.Second*10), onRetry)

		if err != nil {
			i++
			logger.Errorf("Failed to generateForQuestion voice: %v", err)
			continue
		}

		err = writeVoice(voice, sentence, fileName)
		if err != nil {
			log.Errorf("Failed to write voice: %v", err)
		}

		err = writeMetadata(sentence, sentenceDirPath)
		if err != nil {
			log.Errorf("Failed to write metadata %v", err)
		}

		i++
	}

	return nil
}

func writeMetadata(sentence string, dir string) error {
	fileName := fmt.Sprintf("%s/.meta", dir)
	metadata := map[string]string{
		"sentence": sentence,
	}
	contents, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	err = os.WriteFile(fileName, contents, 0777)
	return err
}

// writeVoice writes the voice as a .dca file
func writeVoice(voice []byte, sentence string, filePath string) error {
	log.Infof("Writing voice for sentence '%s' at '%s'", sentence, filePath)

	buffer := bytes.NewReader(voice)
	session, err := dca.ConvertReader(buffer)
	if err != nil {
		return err
	}
	defer session.Cleanup()

	b, err := io.ReadAll(session)
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, b, 0777)
	if err != nil {
		return err
	}

	return nil
}
