package tts

import (
	"app/env"
	"app/logging"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/avast/retry-go/v4"
	"go.uber.org/zap"
	"io"
	"net/http"
	"sync"
	"time"
)

const maxAttempts = 3

var logger = logging.Get().Named("tts")

type Client struct {
	httpClient *http.Client
	host       string
}

type TextToVoiceRequest struct {
	Text    string  `json:"sentence"`
	Speaker Speaker `json:"speaker"`
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: time.Minute * 1,
		},
		// TODO Add option to configure via discord
		host: env.Env.TTSHost,
	}
}

// PreloadVoices loads all available voices into the cache
func (c *Client) PreloadVoices(ctx context.Context, voices []*TextToVoiceRequest) {
	if len(voices) == 0 {
		return
	}

	limiter := make(chan bool, 5)

	wg := sync.WaitGroup{}
	wg.Add(len(voices))

	for _, voice := range voices {
		limiter <- true

		go func() {
			err := retry.Do(func() error {
				_, err := c.TextToVoice(ctx, voice)
				return err
			}, retry.Context(ctx), retry.Attempts(maxAttempts))

			if err != nil {
				logger.Error("failed to preload voice", zap.Error(err), zap.Any("voice", voice))
			}

			wg.Done()
			<-limiter
		}()

	}
	wg.Wait()
}

func (c *Client) TextToVoice(context context.Context, payload *TextToVoiceRequest) ([]byte, error) {
	url := fmt.Sprintf("%s%s", c.host, "/generate")

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(context, http.MethodPost, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("failed to convert text to voice, status code: %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
