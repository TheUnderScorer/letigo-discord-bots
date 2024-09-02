package main

import (
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
	"golang.org/x/net/context"
	"os"
	"src-go/domain/trivia"
	"src-go/messages"
	"src-go/util"
	"strconv"
	"time"
	openai2 "tools/shared/openai"
	util2 "tools/shared/util"
)

func main() {
	messages.Init()

	err := godotenv.Load("../.env")
	if err != nil {
		log.Warn("error loading .env file", err)
	}

	file, err := os.Create(fmt.Sprintf("questions-enhancer/result/result-%d.json", time.Now().Unix()))
	if err != nil {
		log.Fatal("failed to create file", err)
	}
	defer file.Close()

	key := os.Getenv("OPENAI_API_KEY")
	openaiClient := openai.NewClient(key)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*20))
	defer cancel()

	var threadMsg []openai.ThreadMessage
	questions := util2.ReadStepResult[trivia.Question]("questions-grammar")

	for i, q := range questions {
		qStr, err := json.Marshal(q)
		if err != nil {
			log.Error("failed to marshal question", "question", q.Question, "err", err)
		}

		m := fmt.Sprintf("Consider following JSON object with a trivia question: \n %s \n add additional \"incorrect_anwser_messages\" similar to existing ones that will be strictly related to this question grammaticaly.\nIn addition, if possible, include new property \"fun_facts\" with an array fun-facts related to this question (if exists). You can just one record to if if there are not much fun facts. Return only JSON.", string(qStr))

		threadMsg = append(threadMsg, openai.ThreadMessage{
			Content: m,
			Metadata: map[string]any{
				"question":      q.Question,
				"initialPhrase": q.Question,
				"i":             strconv.Itoa(i),
			},
			Role: openai.ThreadMessageRoleUser,
		},
		)
	}

	threadIds := make(map[string]bool)
	result := openai2.SendChunked(ctx, openaiClient, threadMsg)

	for _, msg := range result {
		threadIds[msg.ThreadID] = true
	}

	for tid := range threadIds {
		userMessage, uok := util.Find(result, func(m openai.Message) bool {
			return m.ThreadID == tid && m.Role == openai.ChatMessageRoleUser
		})
		assistantMessage, aok := util.Find(result, func(m openai.Message) bool {
			return m.ThreadID == tid && m.Role == openai.ChatMessageRoleAssistant
		})

		if !uok || !aok {
			continue
		}

		i := userMessage.Metadata["i"].(string)
		iInt, err := strconv.Atoi(i)
		if err != nil {
			log.Error("failed to convert i", "i", i, "err", err)
		}

		var q trivia.Question
		err = json.Unmarshal([]byte(assistantMessage.Content[0].Text.Value), &q)
		if err != nil {
			log.Error("failed to unmarshal question", "question", userMessage.Content, "err", err)
		}

		questions[iInt] = &q
	}

	err = json.NewEncoder(file).Encode(questions)
	if err != nil {
		log.Fatal("failed to write result", err)
	}
}
