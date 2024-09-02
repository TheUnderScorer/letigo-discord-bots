package openai

import (
	"context"
	"errors"
	"github.com/charmbracelet/log"
	"github.com/sashabaranov/go-openai"
	"strings"
	"sync"
	"time"
	"tools/shared/constants"
)

func SendChunked(ctx context.Context, client *openai.Client, messages []openai.ThreadMessage) []openai.Message {
	var out []openai.Message

	limit := make(chan bool, 10)
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(len(messages))

	for i, msg := range messages {
		go func() {
			limit <- true
			defer wg.Done()
			defer func() {
				<-limit
			}()
			log.Infof("Sending %d/%d", i+1, len(messages))

			result, err := SendMessages(ctx, client, []openai.ThreadMessage{msg})
			if err != nil {
				log.Error("Failed to send message chunk", "i", i, "err", err)

				return
			}

			mu.Lock()
			defer mu.Unlock()
			out = append(out, result.Messages...)
		}()
	}

	wg.Wait()

	return out
}

func SendMessages(ctx context.Context, client *openai.Client, messages []openai.ThreadMessage) (*openai.MessagesList, error) {
	result, err := doSend(ctx, client, messages)
	if err != nil {
		if isRateLimitErr(err) {
			log.Warn("rate limit exceeded", "messages", messages)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()

			case <-time.After(time.Second * 5):
				return SendMessages(ctx, client, messages)
			}
		}

		return nil, err
	}

	return result, err
}

func isRateLimitErr(err error) bool {
	if strings.Contains(strings.ToLower(err.Error()), "Rate limit reached") {
		return true
	}

	var e *openai.APIError
	if errors.As(err, &e) {
		if runError, ok := e.Code.(string); ok {
			if runError == string(openai.RunErrorRateLimitExceeded) {
				return true
			}
		}
	}

	return false
}

func doSend(ctx context.Context, client *openai.Client, messages []openai.ThreadMessage) (*openai.MessagesList, error) {
	thread, err := client.CreateThread(ctx, openai.ThreadRequest{
		Messages: messages,
	})
	if err != nil {
		return nil, err
	}

	run, err := client.CreateRun(context.Background(), thread.ID, openai.RunRequest{
		AssistantID: constants.AssistantId,
	})
	if err != nil {
		return nil, err
	}

	err = waitForThread(ctx, run.ThreadID, run.ID, client)
	if err != nil {
		return nil, err
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
