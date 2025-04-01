package llm

import (
	"app/errors"
	messages2 "app/messages"
	openai2 "app/openai"
	"app/util/arrayutil"
	"context"
	"github.com/openai/openai-go"
)

type OpenAIModelDefinition struct {
	Model         openai.ChatModel
	ContextWindow int32
	Encoding      string
}

type OpenAIAdapter struct {
	client *openai.Client
	// openAI model to use
	model OpenAIModelDefinition
}

// NewOpenAIAdapter creates a new instance of OpenAIAdapter with the provided client and model.
func NewOpenAIAdapter(client *openai.Client, model OpenAIModelDefinition) *OpenAIAdapter {
	return &OpenAIAdapter{
		client: client,
		model:  model,
	}
}

func (o *OpenAIAdapter) Prompt(ctx context.Context, p Prompt) (string, error) {
	var messages []*ChatMessage

	if p.Traits != "" {
		messages = append(messages, &ChatMessage{
			Role:     ChatRoleSystem,
			Contents: p.Traits,
		})
	}

	messages = append(messages, &ChatMessage{
		Role:     ChatRoleUser,
		Contents: p.Phrase,
	})

	message, err := o.Chat(ctx, &Chat{
		Messages: messages,
	})
	if err != nil {
		return "", err
	}

	return message.Contents, nil
}

func (o *OpenAIAdapter) Chat(ctx context.Context, chat *Chat) (*ChatMessage, error) {
	var messages []openai.ChatCompletionMessageParamUnion

	for _, chatMessage := range chat.Messages {
		switch chatMessage.Role {
		case ChatRoleUser:
			messages = append(messages, openai.UserMessage(chatMessage.Contents))

		case ChatRoleAssistant:
			messages = append(messages, openai.AssistantMessage(chatMessage.Contents))

		case ChatRoleSystem:
			messages = append(messages, openai.SystemMessage(chatMessage.Contents))
		}
	}

	param := openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    o.model.Model,
	}

	tokens, err := o.countTokens(chat)
	if err != nil {
		return nil, errors.Wrap(err, "failed to count tokens")
	}

	if tokens > o.model.ContextWindow {
		return nil, NewPromptTooLongError(tokens)
	}

	completion, err := o.client.Chat.Completions.New(ctx, param)
	if err != nil {
		return nil, err
	}

	message := completion.Choices[0].Message
	if message.Refusal != "" {
		lastMessage, lastMessageOk := arrayutil.Last(chat.Messages)
		userFriendlyError := errors.NewPublicError(arrayutil.RandomElement(messages2.Messages.Chat.RefuseToReply))
		userFriendlyError.AddContext("refusal", message.Refusal)

		if lastMessageOk {
			userFriendlyError.AddContext("prompt", lastMessage.Contents)
		}

		return nil, userFriendlyError
	}

	return NewChatMessage(message.Content, ChatRoleUser), nil
}

func (o *OpenAIAdapter) countTokens(chat *Chat) (int32, error) {
	var contents string
	for _, message := range chat.Messages {
		contents = contents + message.Contents
	}

	return openai2.CountTokens(contents, o.model.Encoding)
}
