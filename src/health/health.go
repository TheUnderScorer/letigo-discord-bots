package health

import (
	gptserverclient "app/llm"
	"app/logging"
	"context"
	"go.uber.org/zap"
)

func CheckApis(ctx context.Context) {
	log := logging.Get().Named("health-check")

	gptClient := ctx.Value(gptserverclient.GtpServerClientContextKey).(*gptserverclient.API)

	ok, err := gptClient.Status()
	if err != nil {
		log.Fatal("Error checking gpt server status", zap.Error(err))
	}

	if !ok {
		log.Fatal("gpt server status is false")
	}
}
