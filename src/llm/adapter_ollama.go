package llm

import (
	"app/logging"
	"context"
	ollama "github.com/ollama/ollama/api"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"strings"
)

var log = logging.Get().Named("llm").Named("adapter_ollama")

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

func (o *OllamaAdapter) Chat(ctx context.Context, request *Chat) (*ChatMessage, *ChatReplyMetadata, error) {
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
		return nil, nil, err
	}

	log.Debug("got response from llm", zap.Any("request", request), zap.String("model", o.model), zap.Any("response", handler))

	return &ChatMessage{
		Role:     ChatRoleAssistant,
		Contents: strings.TrimSpace(strings.Join(handler.MessageParts, "")),
	}, nil, nil
}

func (o *OllamaAdapter) Prompt(ctx context.Context, p Prompt) (string, *PromptReplyMetadata, error) {
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
		return "", nil, err
	}

	log.Debug("got response from llm", zap.String("prompt", p.Phrase), zap.String("model", o.model), zap.String("system", p.Traits), zap.Any("response", handler))

	return strings.TrimSpace(strings.Join(handler.MessageParts, "")), nil, nil
}

type streamHandler struct {
	MessageParts  []string `json:"message_parts"`
	ThinkingParts []string `json:"thinking_parts"`
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
