package main

import (
	"app/aws"
	"app/bots"
	"app/domain"
	"app/domain/interaction"
	"app/domain/player"
	"app/domain/scheduler"
	"app/domain/trivia"
	"app/domain/tts"
	"app/env"
	"app/logging"
	"app/messages"
	"app/metadata"
	"app/server"
	"context"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

var log = logging.Get().Named("server")

func main() {
	log.Info("Booting app")
	log.Info("App version", zap.String("version", metadata.GetVersion()))

	err := godotenv.Load()
	if err != nil {
		log.Warn("error loading .env file", zap.Error(err))
	}

	env.Init()
	messages.Init()

	app := gin.New()

	if env.IsProd() {
		gin.SetMode(gin.ReleaseMode)
	}

	ttsClient := tts.NewClient()
	cfg, err := aws.NewConfig(context.Background())
	if err != nil {
		log.Fatal("failed to create aws config", zap.Error(err))
	}
	s3Client := s3.NewFromConfig(cfg)

	ctx := context.WithValue(context.Background(), player.ChannelPlayerContextKey, player.NewChannelPlayerManager())
	ctx = context.WithValue(ctx, trivia.ManagerContextKey, trivia.NewManager(ttsClient))
	ctx = context.WithValue(ctx, aws.S3ContextKey, aws.NewS3(s3Client))

	// Bots
	ctx = context.WithValue(ctx, bots.BotNameWojciech, bots.NewBot(bots.BotNameWojciech, env.Env.WojciechBotToken))
	ctx = context.WithValue(ctx, bots.BotNameTadeuszSznuk, bots.NewBot(bots.BotNameTadeuszSznuk, env.Env.TadeuszBotToken))

	app.Use(gin.Recovery())

	server.CreateRouter(ctx, app, metadata.GetVersion())

	go interaction.Init(ctx)
	go domain.Init(ctx)

	err = scheduler.Init(ctx)
	if err != nil {
		log.Fatal("failed to init scheduler", zap.Error(err))
	}

	err = app.Run(":3000")
	if err != nil {
		log.Fatal("failed to start server", zap.Error(err))
	}

}
