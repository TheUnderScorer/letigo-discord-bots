package llm

import (
	"app/logging"
	"app/util/arrayutil"
	"context"
	"encoding/base64"
	ollama "github.com/ollama/ollama/api"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"strings"
)

var log = logging.Get().Named("llm").Named("AdapterOllama")

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
			Images: arrayutil.Map(m.Files, func(file File) ollama.ImageData {
				return file.Data
			}),
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

	log.Info("got response from llm", zap.Any("request", request), zap.String("model", o.model), zap.Strings("response", handler.MessageParts), zap.Strings("thinking", handler.ThinkingParts))

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
		Images: arrayutil.Map(p.Files, func(file File) ollama.ImageData {
			b64 := make([]byte, base64.StdEncoding.EncodedLen(len(file.Data)))
			base64.StdEncoding.Encode(b64, file.Data)
			return b64
		}),
	}

	handler := newStreamHandler()

	err := o.client.Generate(ctx, req, func(response ollama.GenerateResponse) error {
		return handler.Handle(response.Response)
	})

	if err != nil {
		return "", nil, err
	}

	log.Info("got response from llm", zap.String("prompt", p.Phrase), zap.String("model", o.model), zap.String("system", p.Traits), zap.Strings("response", handler.MessageParts), zap.Strings("thinking", handler.ThinkingParts))

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
		log.Debug("llm started thinking")
		h.isThinking = true
		return nil
	}

	if contents == ThinkingEnd {
		log.Debug("llm stopped thinking")
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
