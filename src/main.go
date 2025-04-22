package main

import (
	"app/aws"
	"app/bots"
	"app/discord"
	"app/domain"
	"app/domain/chat"
	"app/domain/interaction"
	openaidomain "app/domain/openai"
	"app/domain/player"
	"app/domain/scheduler"
	"app/domain/trivia"
	"app/domain/tts"
	"app/env"
	"app/llm"
	"app/logging"
	"app/messages"
	"app/metadata"
	"app/server"
	"context"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/pkoukk/tiktoken-go"
	"go.uber.org/zap"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

var log = logging.Get().Named("server")

func main() {
	log.Info("Booting app")
	log.Info("App version", zap.String("version", metadata.GetVersion()))

	err := godotenv.Load()
	if err != nil {
		log.Warn("error loading .env file", zap.Error(err))
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))

	env.Init()
	messages.Init()

	app := gin.New()

	if env.IsProd() {
		gin.SetMode(gin.ReleaseMode)
	}

	app.Use(gin.Recovery())

	httpClient := &http.Client{
		Timeout: time.Minute * 5,
	}

	ttsClient := tts.NewClient()
	cfg, err := aws.NewConfig(context.Background())
	if err != nil {
		log.Fatal("failed to create aws config", zap.Error(err))
	}
	s3Client := s3.NewFromConfig(cfg)

	channelPlayerManager := player.NewChannelPlayerManager()
	triviaManager := trivia.NewManager(ttsClient)
	awsS3 := aws.NewS3(s3Client)

	// Bots
	wojciechBot := bots.NewBot(bots.BotNameWojciech, env.Env.WojciechBotToken)
	tadeuszBot := bots.NewBot(bots.BotNameTadeuszSznuk, env.Env.TadeuszBotToken)

	// DiscordChat
	ollamaUrl, err := url.Parse(env.Env.OllamaHost)
	if err != nil {
		log.Fatal("failed to parse ollama host", zap.Error(err))
	}

	ollamaAdapter := llm.NewOllamaAdapter(env.Env.OllamaModel, ollamaUrl, httpClient)
	ollamaApi := llm.NewAPI(ollamaAdapter, "ollama")
	openAIClient := openai.NewClient(option.WithAPIKey(env.Env.OpenAIApiKey))
	openAIAssistantDefinition := llm.OpenAIAssistantDefinition{
		ID:            env.Env.OpenAIAssistantID,
		Encoding:      tiktoken.MODEL_O200K_BASE,
		ContextWindow: 128_000,
	}
	openAIAssistantAdapter := llm.NewOpenAIAssistantAdapter(&openAIClient, openAIAssistantDefinition, env.Env.OpenAIAssistantVectorStoreID)
	assistantApi := llm.NewAPI(openAIAssistantAdapter, "openai")

	openAIAdapter := llm.NewOpenAIAdapter(&openAIClient, llm.OpenAIModelDefinition{
		Model:         "gpt-4.1-mini",
		ContextWindow: 128_000,
		Encoding:      tiktoken.MODEL_O200K_BASE,
	}, env.Env.OpenAIAssistantVectorStoreID)
	openAIApi := llm.NewAPI(openAIAdapter, "openai")

	llmContainer := &llm.Container{
		AssistantAPI: assistantApi,
		FreeAPI:      ollamaApi,
		ExpensiveAPI: openAIApi,
	}

	chatManager := chat.NewManager(wojciechBot.Session, llmContainer)

	botsArr := []*bots.Bot{wojciechBot, tadeuszBot}

	server.CreateRouter(&server.RouterContainer{
		Bots: botsArr,
	}, app, metadata.GetVersion())

	openaidomain.Init(&openAIClient, env.Env.OpenAIAssistantVectorStoreID)

	interaction.Init(botsArr)

	domain.Init(&domain.Container{
		ChatManager: chatManager,
		LlmApi:      llmContainer.FreeAPI,
		Bots:        botsArr,
		CommandsContainer: &interaction.CommandsContainer{
			TriviaManager:        triviaManager,
			ChannelPlayerManager: channelPlayerManager,
			S3:                   awsS3,
			ComponentInteractionHandlers: []discord.ComponentInteractionHandler{
				trivia.NewComponentInteractionHandler(triviaManager),
				chat.NewForgetComponentHandler(&openAIClient),
			},
		},
	})

	err = scheduler.Init(wojciechBot)
	if err != nil {
		log.Fatal("failed to init scheduler", zap.Error(err))
	}

	err = app.Run(":3000")
	if err != nil {
		log.Fatal("failed to start server", zap.Error(err))
	}

}
