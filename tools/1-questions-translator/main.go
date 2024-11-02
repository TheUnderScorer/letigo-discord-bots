package main

import (
	"app/domain/trivia"
	"app/logging"
	"app/util"
	"context"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
	"os"
	"time"
	openai2 "tools/shared/openai"
	"tools/shared/opentdb"
)

var logger = logging.Get()

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		logger.Warn("error loading .env file", zap.Error(err))
	}

	file, err := os.Create(fmt.Sprintf("questions-translator/result/result-%d.json", time.Now().Unix()))
	if err != nil {
		logger.Fatal("failed to create file", zap.Error(err))
	}
	defer file.Close()

	key := os.Getenv("OPENAI_API_KEY")
	openaiClient := openai.NewClient(key)

	waitCtx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*2))
	defer cancel()
	messages, err := createMessages()
	if err != nil {
		logger.Fatal("failed to create messages", zap.Error(err))
	}

	result, err := openai2.SendMessages(waitCtx, openaiClient, messages)
	if err != nil {
		logger.Fatal("failed to send messages", zap.Error(err))
	}

	logger.Info("messages", zap.Any("messages", result))

	questions := handleResult(result)

	logger.Info("questions", zap.Any("questions", questions))

	if len(questions) == 0 {
		logger.Fatal("no questions found")
	}

	err = json.NewEncoder(file).Encode(questions)
	if err != nil {
		logger.Fatal("failed to write questions", zap.Error(err))
	}
}

func handleResult(messages *openai.MessagesList) (result []trivia.Question) {
	for _, m := range messages.Messages {
		if m.Role != openai.ChatMessageRoleAssistant {
			continue
		}

		for _, content := range m.Content {
			if content.Text == nil {
				continue
			}

			var out struct {
				Questions []opentdb.Question `json:"questions"`
			}

			err := json.Unmarshal([]byte(content.Text.Value), &out)
			if err != nil {
				logger.Error("failed to unmarshal content", zap.Error(err), zap.Any("content", content.Text.Value))

				continue
			}

			result = append(result, util.Map(out.Questions, func(v opentdb.Question) trivia.Question {
				return v.ToTrivia()
			})...)
		}
	}

	return result
}

func createMessages() (messages []openai.ThreadMessage, err error) {
	const limit = 10
	i := 0

	for limit > i {
		questions, err := opentdb.GetQuestions()
		if err != nil {
			return messages, err
		}

		if len(questions) == 0 {
			continue
		}

		messages = append(messages, openai.ThreadMessage{
			Content: fmt.Sprintf("Translate questions, category and answers (if applicable, don't translate names directly) to Polish. Return an JSON array of object with key 'questions' that will contain an array of translated questions. JSON schema of these questions should match the one initially provided. %s", questions),
			Role:    openai.ChatMessageRoleUser,
		})

		i++
	}

	return messages, err
}
