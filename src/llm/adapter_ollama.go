package llm

import (
	"context"
	ollama "github.com/ollama/ollama/api"
	"net/http"
	"net/url"
	"strings"
)

type OllamaAdapter struct {
	client *ollama.Client
	// an Ollama model to use
	model string
}

func NewOllamaAdapter(model string, base *url.URL, http *http.Client) *OllamaAdapter {
	return &OllamaAdapter{
		client: ollama.NewClient(base, http),
		model:  model,
	}
}

func (o *OllamaAdapter) Chat(ctx context.Context, request *Chat) (*ChatMessage, error) {
	stream := true

	var messages []ollama.Message

	for _, m := range request.Messages {
		messages = append(messages, ollama.Message{
			Content: m.Contents,
			Role:    getOllamaRole(m.Role),
		})
	}

	req := &ollama.ChatRequest{
		Model:    o.model,
		Stream:   &stream,
		Messages: messages,
	}

	var chatMessageParts []string

	err := o.client.Chat(ctx, req, func(response ollama.ChatResponse) error {
		chatMessageParts = append(chatMessageParts, response.Message.Content)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &ChatMessage{
		Role:     ChatRoleAssistant,
		Contents: strings.Join(chatMessageParts, ""),
	}, nil
}

func (o *OllamaAdapter) Prompt(ctx context.Context, p Prompt) (string, error) {
	stream := true

	req := &ollama.GenerateRequest{
		Prompt: p.Phrase,
		Model:  o.model,
		System: p.Traits,
		Stream: &stream,
	}

	var messageParts []string

	err := o.client.Generate(ctx, req, func(response ollama.GenerateResponse) error {
		messageParts = append(messageParts, response.Response)

		return nil
	})

	if err != nil {
		return "", err
	}

	return strings.Join(messageParts, ""), nil
}

func getOllamaRole(role ChatRole) string {
	switch role {
	case ChatRoleSystem:
		return "system"

	case ChatRoleUser:
		return "user"

	case ChatRoleAssistant:
		return "assistant"

	}
	return ""
}
