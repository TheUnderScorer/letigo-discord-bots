package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/avast/retry-go/v4"
	"github.com/charmbracelet/log"
	"github.com/joho/godotenv"
	"os"
	"src-go/dca"
	"src-go/domain/trivia"
	tts2 "src-go/domain/tts"
	"src-go/env"
	"src-go/messages"
	util2 "src-go/util"
	"sync"
	"time"
	"tools/shared/util"
)

func main() {
	messages.Init()
	err := godotenv.Load("../.env")
	if err != nil {
		log.Warn("error loading .env file", err)
	}

	env.Init()

	client := tts2.NewClient()

	var wg sync.WaitGroup
	limit := make(chan bool, 2)

	questions := util.ReadStepResult[trivia.Question]("3-questions-enhancer")
	// TODO Remove
	questions = questions[:3]

	wg.Add(len(questions))

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Hour*1))
	defer cancel()

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
			err := generate(ctx, q, client)
			if err != nil {
				log.Errorf("Failed to generate %d/%d: %v", pos, length, err)
			}
			log.Infof("Generated %d/%d", pos, length)
		}()
	}

	wg.Wait()

}

func generate(ctx context.Context, question *trivia.Question, client *tts2.Client) error {
	input := trivia.NewQuestionSentencesInput(question)
	sentences := input.Sentences()

	questionDirectoryPath := fmt.Sprintf("questions-voice-generator/result/%s", question.ID())
	err := os.MkdirAll(questionDirectoryPath, 0777)
	if err != nil && !os.IsExist(err) {
		return err
	}

	var wg sync.WaitGroup
	limit := make(chan bool, 2)

	for _, sentence := range sentences {
		wg.Add(1)
		go func() {
			limit <- true
			defer func() {
				<-limit
				wg.Done()
			}()
			err := generateSentence(ctx, sentence, questionDirectoryPath, client)
			if err != nil {
				log.Errorf("Failed to generate sentence: %s", err)
			}
		}()
	}

	wg.Wait()

	return nil
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
		if errors.Is(err, os.ErrExist) {
			log.Warnf("sentence already exists: %s", sentence)

			return nil
		}

		return err
	}

	onRetry := retry.OnRetry(func(attempt uint, err error) {
		logger.Warnf("Attempt %d to generate voice failed: %v", attempt, err)
	})

	for i < sentenceCount {
		voice, err := retry.DoWithData(func() ([]byte, error) {
			return client.TextToVoice(ctx, &tts2.TextToVoiceRequest{
				Text:    sentence,
				Speaker: tts2.SpeakerTadeusz,
			})
		}, retry.Context(ctx), retry.Attempts(maxAttempts), retry.Delay(time.Second*5), onRetry)

		if err != nil {
			i++
			logger.Errorf("Failed to generate voice: %v", err)
			continue
		}

		fileName := fmt.Sprintf("%s/%d.dca", sentenceDirPath, i)
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

	buffer := bytes.NewBuffer(voice)
	session, err := dca.ConvertReader(buffer)
	if err != nil {
		return err
	}
	defer session.Cleanup()

	opusFrames, err := dca.ReadSession(session)
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, opusFrames, 0777)
	if err != nil {
		return err
	}

	return nil
}
