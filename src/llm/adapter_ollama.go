package llm

import (
	"context"
	ollama "github.com/ollama/ollama/api"
	"net/http"
	"net/url"
	"strings"
)

const ThinkingStart = "<think>"
const ThinkingEnd = "</think>"

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
			Role:    mapOllamaRole(m.Role),
		})
	}

	req := &ollama.ChatRequest{
		Model:    o.model,
		Stream:   &stream,
		Messages: messages,
	}

	handler := newStreamHandler()

	err := o.client.Chat(ctx, req, func(response ollama.ChatResponse) error {
		return handler.Handle(response.Message.Content)
	})

	if err != nil {
		return nil, err
	}

	return &ChatMessage{
		Role:     ChatRoleAssistant,
		Contents: strings.Join(handler.MessageParts, ""),
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

	handler := newStreamHandler()

	err := o.client.Generate(ctx, req, func(response ollama.GenerateResponse) error {
		return handler.Handle(response.Response)
	})

	if err != nil {
		return "", err
	}

	return strings.Join(handler.MessageParts, ""), nil
}

type streamHandler struct {
	MessageParts  []string
	ThinkingParts []string
	isThinking    bool
}

func newStreamHandler() *streamHandler {
	return &streamHandler{
		MessageParts:  []string{},
		ThinkingParts: []string{},
		isThinking:    false,
	}
}

func (h *streamHandler) Handle(contents string) error {
	if contents == ThinkingStart {
		h.isThinking = true
		return nil
	}

	if contents == ThinkingEnd {
		h.isThinking = false
		return nil
	}

	if h.isThinking {
		h.ThinkingParts = append(h.ThinkingParts, contents)
	} else {
		h.MessageParts = append(h.MessageParts, contents)
	}

	return nil
}

func mapOllamaRole(role ChatRole) string {
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
