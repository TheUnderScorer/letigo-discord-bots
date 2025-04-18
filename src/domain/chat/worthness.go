package chat

import (
	"app/llm"
	"app/random"
	"context"
	"fmt"
	"go.uber.org/zap"
	"math"
	"strconv"
	"time"
)

const MinMessageLength = 25
const BaseProbability = 0.4 // Base probability for messages of MinMessageLength
const MaxProbability = 0.9  // Maximum probability cap for very long messages
const LengthFactor = 0.001  // How quickly probability increases with length

// IsWorthyOfReply determines if a given message warrants a reply based on length, probability, and an LLM-generated evaluation.
func IsWorthyOfReply(llmClient *llm.API, message string) bool {
	if len(message) == 0 {
		chatLog.Warn("message is empty")
		return false
	}

	if len(message) < MinMessageLength {
		chatLog.Info("message is too short", zap.String("message", message))

		return false
	}

	// Calculate probability based on message length
	// Start with BASE_PROBABILITY at MIN_MESSAGE_LENGTH
	// and increase based on LENGTH_FACTOR
	lengthFactor := float64(len(message)-MinMessageLength) * LengthFactor
	probability := BaseProbability + lengthFactor

	// Cap the maximum probability
	probability = math.Min(probability, MaxProbability)

	if !random.ChanceOfTrue(probability) {
		chatLog.Info("skipping message based on random chance",
			zap.Int("messageLength", len(message)),
			zap.Float64("probability", probability))

		return false
	}

	requestCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	response, err := llmClient.Prompt(requestCtx, llm.Prompt{
		Phrase: getPromptPhrase(message),
	})

	if err != nil {
		chatLog.Error("failed to get response", zap.Error(err))
		return false
	}

	result, err := strconv.ParseBool(response.Reply)
	if err != nil {
		chatLog.Error("failed to parse response as bool", zap.Error(err), zap.String("reply", response.Reply))
		return false
	}

	return result
}

func getPromptPhrase(message string) string {
	return fmt.Sprintf("Judge if the following message is interesting enough to reply and have a meaningful discussion. Return ONLY 'true' if it is, otherwise return 'false. Never return anything else. Here's the message: \n\n %s", message)
}
