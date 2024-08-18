package openai

import (
	"context"
	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
	"time"
	"tools/constants"
)

func SendMessages(ctx context.Context, client *openai.Client, messages []openai.ThreadMessage) (*openai.MessagesList, error) {
	thread, err := client.CreateThread(ctx, openai.ThreadRequest{
		Messages: messages,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create thread")
	}

	run, err := client.CreateRun(context.Background(), thread.ID, openai.RunRequest{
		AssistantID: constants.AssistantId,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create run")
	}

	err = waitForThread(ctx, run.ThreadID, run.ID, client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wait for thread")
	}

	result, err := client.ListMessage(context.Background(), thread.ID, nil, nil, nil, nil)

	return &result, err
}

func waitForThread(ctx context.Context, threadId string, runId string, client *openai.Client) error {
	for {
		run, err := client.RetrieveRun(ctx, threadId, runId)
		if err != nil {
			return err
		}

		switch run.Status {
		case openai.RunStatusCompleted:
			return nil

		case openai.RunStatusFailed:
			return errors.New(run.LastError.Message)
		}

		select {
		case <-time.After(time.Second * 5):
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
