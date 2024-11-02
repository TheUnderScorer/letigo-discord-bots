package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"net/http"
	"src-go/aws"
	"src-go/bots"
	"src-go/domain"
	"src-go/domain/interaction"
	"src-go/domain/player"
	"src-go/domain/scheduler"
	"src-go/domain/trivia"
	"src-go/domain/tts"
	"src-go/env"
	"src-go/logging"
	"src-go/messages"
	"src-go/server/responses"
	"time"
)

var version string

var log = logging.Get().Named("server")

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Warn("error loading .env file", zap.Error(err))
	}

	env.Init()
	messages.Init()

	r := gin.New()

	if env.IsProd() {
		gin.SetMode(gin.ReleaseMode)
	}

	r.Use(gin.Recovery())
	r.Use(func(c *gin.Context) {
		start := time.Now()

		c.Next()

		end := time.Now()
		duration := end.Sub(start)

		log.Info("Processed request", zap.String("method", c.Request.Method), zap.String("path", c.Request.URL.Path), zap.Int("status", c.Writer.Status()), zap.Duration("duration", duration))
	})

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, &responses.VersionInfo{
			Version: version,
			Result:  true,
		})
	})

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

	interaction.Init(ctx)
	go domain.Init(ctx)

	err = scheduler.Init(ctx)
	if err != nil {
		log.Fatal("failed to init scheduler", zap.Error(err))
	}

	err = r.Run(":8081")
	if err != nil {
		log.Fatal("failed to start server", zap.Error(err))
	}

}
