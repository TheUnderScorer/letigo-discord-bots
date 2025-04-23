package llm

import (
	openaiutil "app/domain/openai"
	"app/errors"
	"app/logging"
	appmessages "app/messages"
	"app/util/arrayutil"
	"bytes"
	"context"
	_ "embed"
	goerrors "errors"
	"github.com/goccy/go-json"
	"github.com/openai/openai-go"
	"go.uber.org/zap"
	"strings"
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
	assistant     OpenAIAssistantDefinition
	vectorStoreID string
}

// NewOpenAIAssistantAdapter creates a new instance of OpenAIAssistantAdapter with the provided client and assistant.
func NewOpenAIAssistantAdapter(client *openai.Client, assistant OpenAIAssistantDefinition, vectorStoreID string) *OpenAIAssistantAdapter {
	return &OpenAIAssistantAdapter{
		client:        client,
		assistant:     assistant,
		vectorStoreID: vectorStoreID,
	}
}

// TODO Add support for attachments
func (o *OpenAIAssistantAdapter) Prompt(ctx context.Context, p Prompt) (string, *PromptReplyMetadata, error) {
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
		chat.AddMessages(message)
	}
	message, _, err := o.Chat(ctx, chat)
	if err != nil {
		return "", nil, err
	}

	return message.Contents, nil, nil
}

func (o *OpenAIAssistantAdapter) Chat(ctx context.Context, chat *Chat) (*ChatMessage, *ChatReplyMetadata, error) {
	log := logging.Get().Named("AdapterOpenAIAssistant").Named("Chat")

	var thread *openai.Thread
	var err error

	tokens, err := o.countTokens(chat)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to count tokens")
	}
	if tokens > o.assistant.ContextWindow {
		return nil, nil, NewPromptTooLongError(tokens)
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
				return nil, nil, errors.Wrap(err, "failed to list thread message ids")
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
						return nil, nil, errors.Wrap(err, "failed to send message to thread")
					}
				}
			}
		}
	} else {
		thread, err = o.createThread(ctx, chat)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to create thread")
		}

		// Store thread ID in chat metadata so that we can retrieve in in further Chat() calls
		chat.AddMetadata(threadIdMetadataKey, thread.ID)
	}

	if thread == nil {
		return nil, nil, goerrors.New("failed to send message: chat contains no thread")
	}

	schema, err := parseSchema()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to parse schema")
	}

	var streamHistory []any

	var additionalInstructions string
	systemMessages := arrayutil.Filter(chat.Messages, func(m *ChatMessage) bool {
		return m.Role == ChatRoleSystem
	})
	if len(systemMessages) > 0 {
		systemPrompts := arrayutil.Map(systemMessages, func(m *ChatMessage) string {
			return m.Contents
		})
		additionalInstructions = strings.Join(systemPrompts, ", ")
	}

	stream := o.client.Beta.Threads.Runs.NewStreaming(ctx, thread.ID, openai.BetaThreadRunNewParams{
		AssistantID: o.assistant.ID,
		ResponseFormat: openai.AssistantResponseFormatOptionParamOfJSONSchema(openai.ResponseFormatJSONSchemaJSONSchemaParam{
			Name:   "bot_response",
			Strict: openai.Bool(true),
			Schema: schema,
		}),
		AdditionalInstructions: openai.String(additionalInstructions),
		Tools: []openai.AssistantToolUnionParam{
			{
				OfFileSearch: &openai.FileSearchToolParam{
					FileSearch: openai.FileSearchToolFileSearchParam{},
				},
			},
		},
	})
	defer stream.Close()
	for stream.Next() {
		err = stream.Err()
		if err != nil {
			return nil, nil, errors.Wrap(err, "stream run error")
		}
		v := stream.Current()

		streamHistory = append(streamHistory, v)

		if v.Event == threadCompletedEvent {
			if len(v.Data.Content) > 0 {
				log.Debug("stream content", zap.Any("contents", v.Data.Content))
				for _, content := range v.Data.Content {
					refusal := content.AsRefusal()
					text := content.AsText()

					if refusal.Refusal != "" {
						log.Error("assistant refusal", zap.String("refusal", refusal.Refusal))
						publicError := errors.NewErrPublic(arrayutil.RandomElement(appmessages.Messages.Chat.RefuseToReply))
						publicError.AddContext("refusal", refusal.Refusal)

						lastMessage, lastMessageOk := arrayutil.FindLast(chat.Messages, func(v *ChatMessage) bool {
							return v.Role == ChatRoleUser
						})
						if lastMessageOk {
							publicError.AddContext("prompt", lastMessage.Contents)
						}

						return nil, nil, publicError
					}

					if text.Text.Value != "" {
						var threadResponse threadResponse
						err = json.Unmarshal([]byte(text.Text.Value), &threadResponse)
						if err != nil {
							log.Error("failed to unmarshal thread", zap.Error(err), zap.String("content", text.Text.Value))
							return nil, nil, errors.Wrap(err, "failed to unmarshal thread response")
						}

						message := NewAssistantChatMessage(threadResponse.Response, v.Data.RunID)
						return message, &ChatReplyMetadata{
							IsWorthRemembering: threadResponse.IsWorthRemembering,
							IsGoodbye:          threadResponse.IsGoodbye,
						}, nil
					}
				}
			}
		}
	}

	err = stream.Err()
	if err != nil {
		return nil, nil, errors.Wrap(err, "stream run error")
	}

	log.Error("assistant didn't reply", zap.Any("streamHistory", streamHistory))

	return nil, nil, goerrors.New("assistant didn't reply")
}

// sendMessageToThread sends a message to an existing thread and returns the response message or an error if it fails.
func (o *OpenAIAssistantAdapter) sendMessageToThread(ctx context.Context, message *ChatMessage, threadId string) error {
	fileIds, err := o.handleAttachments(ctx, message)
	if err != nil {
		log.Error("failed to handle attachments", zap.Error(err))
	}
	messageMetadata := mapMetadata(message)

	content := openai.BetaThreadMessageNewParamsContentUnion{
		OfArrayOfContentParts: createMessageContentParts(message, fileIds),
	}
	_, err = o.client.Beta.Threads.Messages.New(ctx, threadId, openai.BetaThreadMessageNewParams{
		Content:  content,
		Metadata: messageMetadata,
		Role:     mapRole(message),
	})

	if err != nil {
		return err
	}

	return nil
}

func (o *OpenAIAssistantAdapter) handleAttachments(ctx context.Context, message *ChatMessage) ([]string, error) {
	var fileIds []string
	if len(message.Files) > 0 {
		for _, file := range message.Files {
			// TODO In next release, search existing files first to avoid duplicates
			file := openai.File(bytes.NewBuffer(file.Data), file.Name, file.ContentType)
			uploadedFile, err := o.client.Files.New(ctx, openai.FileNewParams{
				File:    file,
				Purpose: openai.FilePurposeVision,
			})
			if err != nil {
				return nil, errors.Wrap(err, "failed to upload file")
			}

			fileIds = append(fileIds, uploadedFile.ID)
		}
	}
	return fileIds, nil
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

	chatMessages := arrayutil.Filter(chat.Messages, func(message *ChatMessage) bool {
		return message.Role != ChatRoleSystem
	})
	messages := arrayutil.Map(chatMessages, func(m *ChatMessage) openai.BetaThreadNewParamsMessage {
		return o.chatMessageToThreadMessage(ctx, m)
	})
	createdThread, err := o.client.Beta.Threads.New(ctx, openai.BetaThreadNewParams{
		Metadata: metadata,
		Messages: messages,
		ToolResources: openai.BetaThreadNewParamsToolResources{
			FileSearch: openai.BetaThreadNewParamsToolResourcesFileSearch{
				VectorStoreIDs: []string{o.vectorStoreID},
			},
		},
	})

	return createdThread, err
}

func (o *OpenAIAssistantAdapter) chatMessageToThreadMessage(ctx context.Context, message *ChatMessage) openai.BetaThreadNewParamsMessage {
	messageMetadata := mapMetadata(message)
	role := mapRoleToString(message)

	fileIds, err := o.handleAttachments(ctx, message)
	if err != nil {
		log.Error("failed to handle attachments", zap.Error(err))
	}

	contents := createMessageContentParts(message, fileIds)

	return openai.BetaThreadNewParamsMessage{
		Role:     role,
		Metadata: messageMetadata,
		Content: openai.BetaThreadNewParamsMessageContentUnion{
			OfArrayOfContentParts: contents,
		},
	}
}

func (o *OpenAIAssistantAdapter) countTokens(chat *Chat) (int32, error) {
	var contents string
	for _, message := range chat.Messages {
		contents = contents + message.Contents
	}

	return openaiutil.CountTokens(contents, o.assistant.Encoding)
}

func createMessageContentParts(message *ChatMessage, fileIds []string) []openai.MessageContentPartParamUnion {
	var contents []openai.MessageContentPartParamUnion
	contents = append(contents, openai.MessageContentPartParamUnion{
		OfText: &openai.TextContentBlockParam{
			Text: message.ChatMessage(),
		},
	})
	if len(fileIds) > 0 {
		for _, fileId := range fileIds {
			contents = append(contents, openai.MessageContentPartParamUnion{
				OfImageFile: &openai.ImageFileContentBlockParam{
					ImageFile: openai.ImageFileParam{
						FileID: fileId,
					},
				},
			})
		}
	}
	return contents
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
