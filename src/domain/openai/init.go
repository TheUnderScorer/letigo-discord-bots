package openai

import (
	chatevents "app/domain/chat/events"
	"app/errors"
	"app/events"
	"bytes"
	"context"
	"fmt"
	"github.com/openai/openai-go"
	"time"
)

func Init(client *openai.Client, vectorStoreID string) {
	events.Handle(func(ctx context.Context, event chatevents.MemoryDetailsExtracted) error {
		reader := bytes.NewReader([]byte(event.Details))

		now := time.Now()
		file := openai.File(reader, fmt.Sprintf("memory_%d.txt", now.Unix()), "text/plain")

		vectorFile, _, err := Remember(ctx, file, client, vectorStoreID)
		if err != nil {
			return errors.Wrap(err, "openai remember failed")
		}

		return events.Dispatch(ctx, MemoryUpdated{
			DiscordThreadID: event.DiscordThreadID,
			Content:         event.Details,
			VectorFileID:    vectorFile.ID,
		})
	})
}
