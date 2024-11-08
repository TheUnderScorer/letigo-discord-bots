package main

import (
	"app/domain/trivia"
	"app/messages"
	"app/util"
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
	"golang.org/x/net/context"
	"os"
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

	file, err := os.Create(fmt.Sprintf("3-questions-enhancer/result/result-%d.json", time.Now().Unix()))
	if err != nil {
		log.Fatal("failed to create file", err)
	}
	defer file.Close()

	key := os.Getenv("OPENAI_API_KEY")
	openaiClient := openai.NewClient(key)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*20))
	defer cancel()

	var threadMsg []openai.ThreadMessage
	questions := util2.ReadStepResult[trivia.Question]("2-questions-grammar")

	for i, q := range questions {
		qStr, err := json.Marshal(q)
		if err != nil {
			log.Error("failed to marshal question", "question", q.Question, "err", err)
		}

		m := fmt.Sprintf("Consider following JSON object with a trivia question: \n \"%s\" \n add additional \"incorrect_anwser_messages\" similar to existing ones that will be strictly related to this question grammaticaly and can be said if the answer is NOT correct. Do not include other possible choices in the answer. The phrases should include the correct answer and be a bit eloquent and witty. NEVER return an result with empty phrases. In addition, if possible, include new property \"fun_facts\" with an array fun-facts related to this question (if exists). You can just one record to if if there are not much fun facts. Return only JSON.", string(qStr))

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

		for _, msg := range q.IncorrectAnswerMessages {
			log.Infof("Got phrase %s", msg)
		}

		questions[iInt] = &q
	}

	err = json.NewEncoder(file).Encode(questions)
	if err != nil {
		log.Fatal("failed to write result", err)
	}
}
