package openai

import (
	"bytes"
	"context"
	"fmt"
	"github.com/openai/openai-go"
	"lib/errors"
	"lib/events"
	"time"
	"wojciech-bot/chat/events"
)

func Init(client *openai.Client, vectorStoreID string) {
	events.Handle(func(ctx context.Context, event chatevents.MemoryDetailsExtracted) error {
		reader := bytes.NewReader([]byte(event.Details))

		now := time.Now()
		file := openai.File(reader, fmt.Sprintf("memory_%d.txt", now.Unix()), "text/plain")

		vectorFile, openAIFile, err := Remember(ctx, file, client, vectorStoreID)
		if err != nil {
			return errors.Wrap(err, "openai remember failed")
		}

		return events.Dispatch(ctx, MemoryUpdated{
			DiscordThreadID: event.DiscordThreadID,
			Content:         event.Details,
			VectorFileID:    vectorFile.ID,
			FileID:          openAIFile.ID,
		})
	})
}
