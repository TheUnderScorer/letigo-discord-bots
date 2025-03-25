package llm

import "context"

type Adapter interface {
	// Prompt sends given prompt to the LLM
	Prompt(ctx context.Context, p Prompt) (string, error)

	// Chat sends a request with chat to the LLM
	Chat(ctx context.Context, chat *Chat) (*ChatMessage, error)
}
