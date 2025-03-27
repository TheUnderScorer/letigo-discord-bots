package llm

import (
	"app/logging"
	"app/metrics"
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
func (api *API) Chat(ctx context.Context, chat *Chat) (*Chat, *ChatMessage, error) {
	api.logger.Debug("sending chat request", zap.Any("chat", chat))

	measure := metrics.NewMeasure()
	measure.Start()
	response, err := api.adapter.Chat(ctx, chat)
	measure.End()

	api.logger.Debug("chat request finished", zap.Duration("duration", measure.Duration()), zap.Any("response", response))

	if err != nil {
		api.logger.Error("failed to send chat request", zap.Error(err))

		return nil, nil, err
	}

	chat.AddMessage(response)

	return chat, response, nil
}

// Prompt sends a request to the LLM with given prompt
func (api *API) Prompt(ctx context.Context, prompt Prompt) (*PromptResponse, error) {
	api.logger.Debug("sending prompt request", zap.Any("prompt", prompt))

	measure := metrics.NewMeasure()
	measure.Start()
	response, err := api.adapter.Prompt(ctx, prompt)
	measure.End()

	api.logger.Debug("prompt response finished", zap.Any("response", response), zap.Duration("duration", measure.Duration()))

	if err != nil {
		return nil, err
	}

	promptResponse := &PromptResponse{
		Reply:    response,
		Duration: measure.Duration().Seconds(),
	}

	return promptResponse, nil
}
