package llm

import (
	"app/errors"
	"app/logging"
	appmessages "app/messages"
	openai2 "app/openai"
	"app/util/arrayutil"
	"context"
	_ "embed"
	goerrors "errors"
	"github.com/goccy/go-json"
	"github.com/openai/openai-go"
	"go.uber.org/zap"
)

const threadCompletedEvent = "thread.message.completed"

//go:embed assistant_json_schema.json
var schema []byte

func parseSchema() (any, error) {
	var v map[string]any
	err := json.Unmarshal(schema, &v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

const threadIdMetadataKey = "openAIThreadID"

type threadResponse struct {
	// Response is a response from the LLM
	Response string `json:"response"`
	// IsWorthRemembering indicates whether previous message has details that are worth remembering
	IsWorthRemembering bool `json:"is_worth_remembering"`
	// IsGoodbye indicates if previous message is a goodbye, or prompt to end the discussion.
	IsGoodbye bool `json:"is_goodbye"`
}

type OpenAIAssistantDefinition struct {
	ID            string
	ContextWindow int32
	Encoding      string
}

type OpenAIAssistantAdapter struct {
	client *openai.Client
	// openAI assistant to use
	assistant OpenAIAssistantDefinition
}

// NewOpenAIAssistantAdapter creates a new instance of OpenAIAssistantAdapter with the provided client and assistant.
func NewOpenAIAssistantAdapter(client *openai.Client, assistant OpenAIAssistantDefinition) *OpenAIAssistantAdapter {
	return &OpenAIAssistantAdapter{
		client:    client,
		assistant: assistant,
	}
}

func (o *OpenAIAssistantAdapter) Prompt(ctx context.Context, p Prompt) (string, error) {
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

	chat := NewChat()
	for _, message := range messages {
		chat.AddMessage(message)
	}
	message, err := o.Chat(ctx, chat)
	if err != nil {
		return "", err
	}

	return message.Contents, nil
}

func (o *OpenAIAssistantAdapter) Chat(ctx context.Context, chat *Chat) (*ChatMessage, error) {
	log := logging.Get().Named("AdapterOpenAIAssistant").Named("Chat")

	var thread *openai.Thread
	var err error

	tokens, err := o.countTokens(chat)
	if err != nil {
		return nil, errors.Wrap(err, "failed to count tokens")
	}
	if tokens > o.assistant.ContextWindow {
		return nil, NewPromptTooLongError(tokens)
	}

	if resolvedThreadId, ok := chat.Metadata[threadIdMetadataKey]; ok {
		// Thread already exists, try to retrieve it
		getThread, err := o.client.Beta.Threads.Get(ctx, resolvedThreadId)
		if err != nil {
			log.Error("failed to get thread", zap.String("threadId", resolvedThreadId), zap.Error(err))
		} else {
			thread = getThread

			threadMessageIds, err := o.listMessageIds(ctx, resolvedThreadId)
			if err != nil {
				return nil, errors.Wrap(err, "failed to list thread message ids")
			}

			// Filter for messages sent by users, not us
			userMessages := arrayutil.Filter(chat.Messages, func(m *ChatMessage) bool {
				return m.Role == ChatRoleUser
			})

			// Find chat messages that were not sent to thread and send them
			for _, chatMessage := range userMessages {
				if chatMessage.ID == "" || !arrayutil.Includes(threadMessageIds, chatMessage.ID) {
					err = o.sendMessageToThread(ctx, chatMessage, resolvedThreadId)
					if err != nil {
						return nil, errors.Wrap(err, "failed to send message to thread")
					}
				}
			}
		}
	} else {
		thread, err = o.createThread(ctx, chat)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create thread")
		}

		// Store thread ID in chat metadata so that we can retrieve in in further Chat() calls
		chat.AddMetadata(threadIdMetadataKey, thread.ID)
	}

	if thread == nil {
		return nil, goerrors.New("failed to send message: chat contains no thread")
	}

	schema, err := parseSchema()
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse schema")
	}

	var streamHistory []any

	stream := o.client.Beta.Threads.Runs.NewStreaming(ctx, thread.ID, openai.BetaThreadRunNewParams{
		AssistantID: o.assistant.ID,
		ResponseFormat: openai.AssistantResponseFormatOptionParamOfJSONSchema(openai.ResponseFormatJSONSchemaJSONSchemaParam{
			Name:   "bot_response",
			Strict: openai.Bool(true),
			Schema: schema,
		}),
	})
	defer stream.Close()
	for stream.Next() {
		err = stream.Err()
		if err != nil {
			return nil, errors.Wrap(err, "stream run error")
		}
		v := stream.Current()

		streamHistory = append(streamHistory, v)

		if v.Event == threadCompletedEvent {
			if v.Data.Content != nil && len(v.Data.Content) > 0 {
				log.Debug("stream content", zap.Any("contents", v.Data.Content))
				for _, content := range v.Data.Content {
					refusal := content.AsRefusal()
					text := content.AsText()

					if refusal.Refusal != "" {
						log.Error("assistant refusal", zap.String("refusal", refusal.Refusal))
						publicError := errors.NewPublicError(arrayutil.RandomElement(appmessages.Messages.Chat.RefuseToReply))
						publicError.AddContext("refusal", refusal.Refusal)

						lastMessage, lastMessageOk := arrayutil.FindLast(chat.Messages, func(v *ChatMessage) bool {
							return v.Role == ChatRoleUser
						})
						if lastMessageOk {
							publicError.AddContext("prompt", lastMessage.Contents)
						}

						return nil, publicError
					}

					if text.Text.Value != "" {
						var threadResponse threadResponse
						err = json.Unmarshal([]byte(text.Text.Value), &threadResponse)
						if err != nil {
							log.Error("failed to unmarshal thread", zap.Error(err), zap.String("content", text.Text.Value))
							return nil, errors.Wrap(err, "failed to unmarshal thread response")
						}

						// Handle remembering
						if threadResponse.IsWorthRemembering {
							log.Info("previous message is worth remembering")
						}

						if threadResponse.IsGoodbye {
							log.Info("previous message is goodbye")
						}

						message := NewChatMessage(threadResponse.Response, ChatRoleAssistant)
						message.IsGoodbye = threadResponse.IsGoodbye
						return message, nil
					}
				}
			}
		}
	}

	err = stream.Err()
	if err != nil {
		return nil, errors.Wrap(err, "stream run error")
	}

	log.Error("assistant didn't reply", zap.Any("streamHistory", streamHistory))

	return nil, goerrors.New("assistant didn't reply")
}

// sendMessageToThread sends a message to an existing thread and returns the response message or an error if it fails.
func (o *OpenAIAssistantAdapter) sendMessageToThread(ctx context.Context, message *ChatMessage, threadId string) error {
	messageMetadata := mapMetadata(message)

	_, err := o.client.Beta.Threads.Messages.New(ctx, threadId, openai.BetaThreadMessageNewParams{
		Content: openai.BetaThreadMessageNewParamsContentUnion{
			OfString: openai.String(message.Contents),
		},
		Metadata: messageMetadata,
		Role:     mapRole(message),
	})

	if err != nil {
		return err
	}

	return nil
}

// listMessageIds retrieves a list of message IDs from the specified thread using OpenAI client pagination.
func (o *OpenAIAssistantAdapter) listMessageIds(ctx context.Context, threadId string) (result []string, err error) {
	messages := o.client.Beta.Threads.Messages.ListAutoPaging(ctx, threadId, openai.BetaThreadMessageListParams{})
	if messages.Err() != nil {
		return result, messages.Err()
	}

	for messages.Next() {
		currentMessage := messages.Current()
		if id, ok := currentMessage.Metadata["id"]; ok {
			result = append(result, id)
		}
	}

	return result, nil
}

func (o *OpenAIAssistantAdapter) createThread(ctx context.Context, chat *Chat) (*openai.Thread, error) {
	metadata := openai.MetadataParam{}
	for k, v := range chat.Metadata {
		metadata[k] = v
	}

	createdThread, err := o.client.Beta.Threads.New(ctx, openai.BetaThreadNewParams{
		Metadata: metadata,
		Messages: arrayutil.Map(chat.Messages, chatMessageToThreadMessage),
	})

	return createdThread, err
}

func chatMessageToThreadMessage(message *ChatMessage) openai.BetaThreadNewParamsMessage {
	messageMetadata := mapMetadata(message)
	role := mapRoleToString(message)

	return openai.BetaThreadNewParamsMessage{
		Role:     role,
		Metadata: messageMetadata,
		Content: openai.BetaThreadNewParamsMessageContentUnion{
			OfString: openai.String(message.Contents),
		},
	}
}

func mapMetadata(message *ChatMessage) openai.MetadataParam {
	messageMetadata := openai.MetadataParam{}
	messageMetadata["id"] = message.ID
	for k, v := range message.Metadata {
		messageMetadata[k] = v
	}
	return messageMetadata
}

func mapRole(message *ChatMessage) openai.BetaThreadMessageNewParamsRole {
	var role openai.BetaThreadMessageNewParamsRole
	switch message.Role {
	case ChatRoleUser:
		role = openai.BetaThreadMessageNewParamsRoleUser
	default:
		role = openai.BetaThreadMessageNewParamsRoleAssistant
	}

	return role
}

func mapRoleToString(message *ChatMessage) string {
	var role string
	switch message.Role {
	case ChatRoleUser:
		role = "user"
	default:
		role = "assistant"
	}

	return role
}

func (o *OpenAIAssistantAdapter) countTokens(chat *Chat) (int32, error) {
	var contents string
	for _, message := range chat.Messages {
		contents = contents + message.Contents
	}

	return openai2.CountTokens(contents, o.assistant.Encoding)
}
