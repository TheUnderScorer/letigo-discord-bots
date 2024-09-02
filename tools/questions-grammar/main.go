package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
	"os"
	"path/filepath"
	"src-go/domain/trivia"
	"src-go/messages"
	"src-go/util"
	"strconv"
	"sync"
	"time"
	openai2 "tools/shared/openai"
	trivia2 "tools/shared/trivia"
	util2 "tools/shared/util"
)

var mu sync.Mutex

func main() {
	messages.Init()

	err := godotenv.Load("../.env")
	if err != nil {
		log.Warn("error loading .env file", err)
	}

	file, err := os.Create(fmt.Sprintf("questions-grammar/result/result-%d.json", time.Now().Unix()))
	if err != nil {
		log.Fatal("failed to create file", err)
	}
	defer file.Close()

	key := os.Getenv("OPENAI_API_KEY")
	openaiClient := openai.NewClient(key)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*20))
	defer cancel()

	var threadMsg []openai.ThreadMessage
	questions := util2.ReadStepResult[trivia.Question]("questions-translator")

	for i, q := range questions {
		if q.Type == "" {
			trivia2.FigureOutType(q)
		}

		if q.Type != trivia.MultipleChoice {
			continue
		}

		var invalidAnswers []string

		if q.Type == trivia.TrueFalse {
			invalidAnswers = messages.Messages.Trivia.InvalidAnswer.Boolean
		} else {
			invalidAnswers = messages.Messages.Trivia.InvalidAnswer.Multiple
		}

		for _, v := range invalidAnswers {
			phrase := util.ApplyTokens(v, map[string]string{
				"ANSWER": q.Correct,
			})

			log.Info("Preparing phrase", "question", q.Question, "phrase", phrase)

			threadMsg = append(threadMsg, openai.ThreadMessage{
				Content: fmt.Sprintf("Rephrase this sentence to sound correctly in polish. For example, in case of this question: \"Która z tych postaci NIE jest grywalna w grze wideo Overwatch z 2016 roku?\" and this answer: \"Niestety, chodziło o: Invoker\" the correct grammatic form is: \"Niestety, chodziło o Invokera\". It is based on the \"1 z 10\" trivia show. Return JSON with key \"value\".  Now provide this phrase in correct grammatic form:\"%s\" ", phrase),
				Metadata: map[string]any{
					"question":      q.Question,
					"initialPhrase": v,
					"phrase":        phrase,
					"i":             strconv.Itoa(i),
				},
				Role: openai.ThreadMessageRoleUser,
			},
			)
		}

	}

	log.Info("Sending messages", "messages", len(threadMsg))

	var wg sync.WaitGroup
	limit := make(chan bool, 10)

	for i, msg := range threadMsg {
		limit <- true
		wg.Add(1)
		go func() {
			log.Infof("Sending %d/%d", i+1, len(threadMsg))
			handleMessage(ctx, openaiClient, msg, questions)
			log.Infof("Sent %d/%d", i+1, len(threadMsg))
			wg.Done()
			<-limit
		}()
	}

	wg.Wait()

	err = json.NewEncoder(file).Encode(questions)
	if err != nil {
		log.Fatal("failed to write file", err)
	}
}

func handleMessage(ctx context.Context, client *openai.Client, msg openai.ThreadMessage, questions []*trivia.Question) {
	result, err := openai2.SendMessages(ctx, client, []openai.ThreadMessage{msg})

	if err != nil {
		var e *openai.APIError
		if errors.As(err, &e) {
			if runError, ok := e.Code.(string); ok {
				if runError == string(openai.RunErrorRateLimitExceeded) {
					log.Warn("rate limit exceeded", "msg", msg.Content)

					select {
					case <-ctx.Done():
						return

					case <-time.After(time.Second * 5):
						handleMessage(ctx, client, msg, questions)
						return
					}
				}
			}

			return
		}

		log.Fatal("failed to send messages", err)
	}

	var assistantMessage openai.Message
	var userMessage openai.Message

	for _, v := range result.Messages {
		if v.Role == openai.ChatMessageRoleAssistant {
			assistantMessage = v
		} else if v.Role == openai.ChatMessageRoleUser {
			userMessage = v
		}
	}

	if userMessage.Content == nil || assistantMessage.Content == nil {
		log.Error("failed to get content", "userMessage", userMessage, "assistantMessage", assistantMessage)

		return
	}

	i, err := strconv.Atoi(userMessage.Metadata["i"].(string))
	if err != nil {
		log.Error("failed to parse i", err)

		return
	}

	var value struct {
		Value string `json:"value"`
	}

	err = json.Unmarshal([]byte(assistantMessage.Content[0].Text.Value), &value)
	if err != nil {
		log.Error("failed to unmarshal value", err)

		return
	}

	mu.Lock()
	defer mu.Unlock()

	questions[i].IncorrectAnswerMessages = append(questions[i].IncorrectAnswerMessages, value.Value)
}

func readQuestions() (q []*trivia.Question) {
	files, err := filepath.Glob("questions-translator/result/result-*.json")
	if err != nil {
		log.Fatal("failed to read files", err)
	}

	for _, fileName := range files {
		contents, err := os.ReadFile(fileName)
		if err != nil {
			log.Error("failed to read file", err)

			continue
		}

		var v []*trivia.Question
		err = json.Unmarshal(contents, &v)
		if err != nil {
			log.Error("failed to unmarshal file", err)

			continue
		}

		q = append(q, v...)
	}

	return q
}
