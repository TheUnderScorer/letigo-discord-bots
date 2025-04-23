package llm

import (
	"app/errors"
	"app/logging"
	"app/messages"
	"app/metrics"
	"app/util/arrayutil"
	"context"
	"go.uber.org/zap"
)

var logger = logging.Get().Named("llm").Named("client")

type API struct {
	adapter Adapter
	name    string
	logger  *zap.Logger
}

type PromptResponse struct {
	Reply    string  `json:"reply"`
	Duration float64 `json:"duration"`
}

func NewAPI(adapter Adapter, name string) *API {
	return &API{
		adapter: adapter,
		logger:  logger.Named(name),
		name:    name,
	}
}

// Chat creates, or continues given chat discussion between user and the assistant (llm model)
func (api *API) Chat(ctx context.Context, chat *Chat) (*Chat, *ChatMessage, *ChatReplyMetadata, error) {
	api.logger.Info("sending chat request", zap.Any("chat", chat))

	measure := metrics.NewMeasure()
	measure.Start()
	response, metadata, err := api.adapter.Chat(ctx, chat)
	measure.End()

	api.logger.Info("chat request finished", zap.Duration("duration", measure.Duration()), zap.Any("response", response))

	if err != nil {
		api.logger.Error("failed to send chat request", zap.Error(err))

		publicErr := errors.NewErrPublicCause(arrayutil.RandomElement(messages.Messages.Chat.FailedToReply), err)

		return nil, nil, nil, publicErr
	}

	chat.AddMessages(response)

	return chat, response, metadata, nil
}

// Prompt sends a request to the LLM with the given prompt
func (api *API) Prompt(ctx context.Context, prompt Prompt) (*PromptResponse, *PromptReplyMetadata, error) {
	log := api.logger.With(zap.String("prompt", prompt.Phrase), zap.String("traits", prompt.Traits), zap.Bool("hasFiles", len(prompt.Files) > 0))
	log.Info("sending prompt request")

	measure := metrics.NewMeasure()
	measure.Start()
	response, metadata, err := api.adapter.Prompt(ctx, prompt)
	measure.End()

	log.Info("prompt response finished", zap.Any("response", response), zap.Duration("duration", measure.Duration()))

	if err != nil {
		publicErr := errors.NewErrPublicCause(arrayutil.RandomElement(messages.Messages.Chat.FailedToReply), err)

		return nil, metadata, publicErr
	}

	promptResponse := &PromptResponse{
		Reply:    response,
		Duration: measure.Duration().Seconds(),
	}

	return promptResponse, metadata, nil
}
