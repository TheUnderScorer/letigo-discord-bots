package llm

import (
	openai2 "app/domain/openai"
	"app/errors"
	messages2 "app/messages"
	"app/util/arrayutil"
	"context"
	goerrors "errors"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/responses"
)

var AssistantDidNotReplyError = goerrors.New("assistant didn't reply")

type OpenAIModelDefinition struct {
	Model         openai.ChatModel
	ContextWindow int32
	Encoding      string
}

type OpenAIAdapter struct {
	client *openai.Client
	// openAI model to use
	model         OpenAIModelDefinition
	vectorStoreID string
}

// NewOpenAIAdapter creates a new instance of OpenAIAdapter with the provided client and model.
func NewOpenAIAdapter(client *openai.Client, model OpenAIModelDefinition, vectorStoreID string) *OpenAIAdapter {
	return &OpenAIAdapter{
		client:        client,
		model:         model,
		vectorStoreID: vectorStoreID,
	}
}

func (o *OpenAIAdapter) Prompt(ctx context.Context, p Prompt) (string, *PromptReplyMetadata, error) {
	res, err := o.client.Responses.New(ctx, responses.ResponseNewParams{
		Model: o.model.Model,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(p.Phrase),
		},
		Instructions: openai.String(p.Traits),
		Tools: []responses.ToolUnionParam{
			{
				OfFileSearch: &responses.FileSearchToolParam{
					VectorStoreIDs: []string{
						o.vectorStoreID,
					},
				},
			},
		},
	})

	if err != nil {
		return "", nil, errors.Wrap(err, "failed to create response")
	}

	metadata := PromptReplyMetadata{}

	for _, out := range res.Output {
		asMessage := out.AsMessage()

		if asMessage.Role == "assistant" {
			for _, contentUnion := range asMessage.Content {
				asOutputText := contentUnion.AsOutputText()

				if len(asOutputText.Annotations) > 0 {
					for _, annotation := range asOutputText.Annotations {
						if annotation.FileID != "" {
							hasMemoryReference := true
							metadata.HasMemoryReference = &hasMemoryReference
						}
					}
				}

				if asOutputText.Text != "" {
					return asOutputText.Text, &metadata, nil
				}
			}
		}
	}

	return "", nil, AssistantDidNotReplyError
}

func (o *OpenAIAdapter) Chat(ctx context.Context, chat *Chat) (*ChatMessage, *ChatReplyMetadata, error) {
	var messages []openai.ChatCompletionMessageParamUnion

	for _, chatMessage := range chat.Messages {
		switch chatMessage.Role {
		case ChatRoleUser:
			messages = append(messages, openai.UserMessage(chatMessage.ChatMessage()))

		case ChatRoleAssistant:
			messages = append(messages, openai.AssistantMessage(chatMessage.ChatMessage()))

		case ChatRoleSystem:
			messages = append(messages, openai.SystemMessage(chatMessage.ChatMessage()))
		}
	}

	param := openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    o.model.Model,
		Tools:    []openai.ChatCompletionToolParam{},
	}

	tokens, err := o.countTokens(chat)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to count tokens")
	}

	if tokens > o.model.ContextWindow {
		return nil, nil, NewPromptTooLongError(tokens)
	}

	completion, err := o.client.Chat.Completions.New(ctx, param)
	if err != nil {
		return nil, nil, err
	}

	message := completion.Choices[0].Message
	if message.Refusal != "" {
		lastMessage, lastMessageOk := arrayutil.Last(chat.Messages)
		userFriendlyError := errors.NewPublicError(arrayutil.RandomElement(messages2.Messages.Chat.RefuseToReply))
		userFriendlyError.AddContext("refusal", message.Refusal)

		if lastMessageOk {
			userFriendlyError.AddContext("prompt", lastMessage.Contents)
		}

		return nil, nil, userFriendlyError
	}

	return NewChatMessage(message.Content, ChatRoleUser), nil, nil
}

func (o *OpenAIAdapter) countTokens(chat *Chat) (int32, error) {
	var contents string
	for _, message := range chat.Messages {
		contents = contents + message.Contents
	}

	return openai2.CountTokens(contents, o.model.Encoding)
}
