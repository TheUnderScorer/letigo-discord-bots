package openai

import (
	"bytes"
	"context"
	"fmt"
	"github.com/openai/openai-go"
	"go.uber.org/zap"
	"io"
	"lib/logging"
	"time"
)

var log = logging.Get().Named("openai").Named("memory")

const system = "You are a helpful assistant that parses messages, and extracts details from them. Extract details from this message someone else should remember, and return ONLY them. Try to answer following questions: WHO? and WHAT? Return single sentence that answers these both questions. If message does not contain anything relevant, or you are enable to answer BOTH of these questions, return an empty string."

func ParseAndRememberText(ctx context.Context, message openai.ChatCompletionMessageParamUnion, client *openai.Client, vectorStoreID string) (*openai.VectorStoreFile, *openai.FileObject, error) {
	localLog := log.With(zap.Any("message", message))
	localLog.Debug("about to remember")

	response, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.ChatModelGPT4oMini,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(system),
			message,
		},
	})
	if err != nil {
		localLog.Error("failed to remember", zap.Any("message", message), zap.Error(err))
		return nil, nil, err
	}
	reply := response.Choices[0].Message
	if reply.Refusal != "" {
		return nil, nil, fmt.Errorf("openai refused to reply to prompt to parse message: %s", reply.Refusal)
	}
	replyBytes := []byte(reply.Content)

	now := time.Now()
	file := openai.File(bytes.NewBuffer(replyBytes), fmt.Sprintf("memory_%d.txt", now.Unix()), "text/plain")

	return Remember(ctx, file, client, vectorStoreID)
}

func Remember(ctx context.Context, contents io.Reader, client *openai.Client, vectorStoreID string) (*openai.VectorStoreFile, *openai.FileObject, error) {
	file, err := client.Files.New(ctx, openai.FileNewParams{
		File:    contents,
		Purpose: openai.FilePurposeAssistants,
	})
	if err != nil {
		return nil, nil, err
	}

	vectorFile, err := client.VectorStores.Files.New(ctx, vectorStoreID, openai.VectorStoreFileNewParams{
		FileID: file.ID,
	})
	if err != nil {
		return nil, nil, err
	}

	log.Info("created new vector file", zap.String("fileID", file.ID), zap.String("vectorFileID", vectorFile.ID))

	return vectorFile, file, nil
}
