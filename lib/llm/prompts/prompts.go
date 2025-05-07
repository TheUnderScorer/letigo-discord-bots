package prompts

import (
	"context"
	"fmt"
	"lib/llm"
	"strconv"
)

// SummarizeDiscordThread generates a short summary in Polish for a given Discord message using an LLM API.
// Returns the summarized prompt response or an error if the operation fails.
func SummarizeDiscordThread(ctx context.Context, llmAPI *llm.API, messageContent string) (*llm.PromptResponse, error) {
	reply, _, err := llmAPI.Prompt(ctx, llm.Prompt{
		Phrase: fmt.Sprintf("summarize this message in short sentence (up to 100 characters) in polish. Make it a neutral sentence, without any special discord syntax (such as mentions, etc.): %s", messageContent),
	})

	return reply, err
}

// IsMessageGoodbye determines if the provided message indicates a goodbye or the end of a discussion session.
func IsMessageGoodbye(ctx context.Context, llmAPI *llm.API, messageContent string) (bool, error) {
	result, _, err := llmAPI.Prompt(ctx, llm.Prompt{
		Phrase: fmt.Sprintf("check, if this message is a goodbye, or prompt to end the discussion. Return ONLY true, if it is, or false otherwise: \n %s", messageContent),
	})
	if err != nil {
		return false, err
	}

	b, err := strconv.ParseBool(result.Reply)
	if err != nil {
		return false, err
	}

	return b, nil
}
