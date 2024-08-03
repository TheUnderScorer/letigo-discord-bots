package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"net/http"
	"src-go/discord"
	"src-go/domain"
	"src-go/domain/interaction"
	"src-go/domain/player"
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

	ctx := context.WithValue(context.Background(), player.ChannelPlayerContextKey, player.NewChannelPlayerManager())

	discordClient := discord.NewClient(env.Cfg.BotToken)
	interaction.Init(discordClient)
	go domain.Init(discordClient, ctx)

	err = r.Run()
	if err != nil {
		log.Fatal("failed to start server", zap.Error(err))
	}

}
