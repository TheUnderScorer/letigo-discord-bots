package chat

import (
	"app/env"
	"app/llm"
	"app/random"
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"math"
	"strconv"
)

const MinMessageLength = 25
const BaseProbability = 0.4 // Base probability for messages of MinMessageLength
const MaxProbability = 0.9  // Maximum probability cap for very long messages
const LengthFactor = 0.001  // How quickly probability increases with length

// IsWorthyOfReply determines if a given message warrants a reply based on length, probability, and an LLM-generated evaluation.
func IsWorthyOfReply(ctx context.Context, llmClient *llm.API, message *discordgo.Message) bool {
	if env.Env.AreAllMessagesReplyWorthy() {
		chatLog.Info("message is worthy of reply because env.AreAllMessagesReplyWorthy is true")
		return true
	}

	hasAttachments := len(message.Attachments) > 0
	contentLength := len(message.Content)
	if contentLength == 0 && hasAttachments {
		chatLog.Warn("message is empty")
		return false
	}

	if contentLength < MinMessageLength && !hasAttachments {
		chatLog.Info("messageContent is too short", zap.String("messageContent", message.Content))

		return false
	}

	if contentLength > 0 {
		// Calculate probability based on messageContent length
		// Start with BASE_PROBABILITY at MIN_MESSAGE_LENGTH
		// and increase based on LENGTH_FACTOR
		lengthFactor := float64(contentLength-MinMessageLength) * LengthFactor
		probability := BaseProbability + lengthFactor

		// Cap the maximum probability
		probability = math.Min(probability, MaxProbability)

		// TODO Remove env check?
		if !random.ChanceOfTrue(probability) && env.IsProd() {
			chatLog.Info("skipping messageContent based on random chance",
				zap.Int("messageLength", contentLength),
				zap.Float64("probability", probability))

			return false
		}
	}

	response, _, err := llmClient.Prompt(ctx, llm.Prompt{
		Phrase: getPromptPhrase(message.Content),
		Files:  llm.HandleDiscordMessageAttachments(message),
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
	return fmt.Sprintf("Judge if the following message is interesting enough to reply and have a meaningful discussion. Take into account files attached to it. Return ONLY 'true' if it is, otherwise return 'false. Never return anything else. Here's the message: \n\n %s", message)
}
