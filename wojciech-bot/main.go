package main

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/pkoukk/tiktoken-go"
	"go.uber.org/zap"
	"lib/discord"
	libenv "lib/env"
	"lib/events"
	libllm "lib/llm"
	"lib/logging"
	"lib/metadata"
	"lib/server"
	"net/http"
	"net/url"
	"time"
	"wojciech-bot/chat"
	"wojciech-bot/env"
	"wojciech-bot/messages"
	openaidomain "wojciech-bot/openai"
	"wojciech-bot/player"
	"wojciech-bot/scheduler"
)

var log = logging.Get().Named("wojciech-bot")

func main() {
	log.Info("Booting app...")
	log.Info("App version", zap.String("version", metadata.GetVersion()))
	err := godotenv.Load()
	if err != nil {
		log.Warn("Failed to load .env file", zap.Error(err))
	}

	env.Init()
	messages.Init()

	bot := discord.NewBot(env.Env.BotToken, "Wojciech", discord.BotMessages{
		UnknownError: messages.Messages.UnknownError,
	})

	app := gin.New()

	if libenv.IsProd() {
		gin.SetMode(gin.ReleaseMode)
	}

	app.Use(gin.Recovery())

	httpClient := &http.Client{
		Timeout: time.Minute * 5,
	}

	// Player
	channelPlayerManager := player.NewChannelPlayerManager()
	playerDomain := player.NewInteractions(channelPlayerManager, bot)

	// DiscordChat
	ollamaUrl, err := url.Parse(env.Env.OllamaHost)
	if err != nil {
		log.Fatal("failed to parse ollama host", zap.Error(err))
	}

	ollamaAdapter := libllm.NewOllamaAdapter(env.Env.OllamaModel, ollamaUrl, httpClient)
	if env.Env.OllamaVisionModel != "" {
		ollamaAdapter.WithVision(env.Env.OllamaVisionModel)
	}

	ollamaApi := libllm.NewAPI(ollamaAdapter, "ollama")
	openAIClient := openai.NewClient(option.WithAPIKey(env.Env.OpenAIApiKey))
	openAIAssistantDefinition := libllm.OpenAIAssistantDefinition{
		ID:            env.Env.OpenAIAssistantID,
		Encoding:      tiktoken.MODEL_O200K_BASE,
		ContextWindow: 128_000,
	}
	openAIAssistantAdapter := libllm.NewOpenAIAssistantAdapter(&openAIClient, openAIAssistantDefinition, env.Env.OpenAIAssistantVectorStoreID)
	assistantApi := libllm.NewAPI(openAIAssistantAdapter, "openai")

	openAIAdapter := libllm.NewOpenAIAdapter(&openAIClient, libllm.OpenAIModelDefinition{
		Model:         "gpt-4.1-mini",
		ContextWindow: 128_000,
		Encoding:      tiktoken.MODEL_O200K_BASE,
	}, env.Env.OpenAIAssistantVectorStoreID)
	openAIApi := libllm.NewAPI(openAIAdapter, "openai")

	llmContainer := &libllm.Container{
		AssistantAPI: assistantApi,
		FreeAPI:      ollamaApi,
		ExpensiveAPI: openAIApi,
	}

	chatManager := chat.NewManager(bot, llmContainer)
	chatScanner := chat.NewDiscordChannelScanner(bot, llmContainer.FreeAPI, func(message *discordgo.Message) {
		err := bot.MessageReactionAdd(message.ChannelID, message.ID, discord.ReactionSeen)
		if err != nil {
			log.Error("failed to add seen reaction", zap.Error(err))
		}

		discordChat := chatManager.GetOrCreateChat(message.ChannelID)
		err = discordChat.HandleNewMessage(message)
		if err != nil {
			log.Error("failed to handle new message", zap.Error(err))
		}
	})
	go chatScanner.Start()
	events.Handle(func(ctx context.Context, event openaidomain.MemoryUpdated) error {
		if event.DiscordThreadID != "" {
			return chat.HandleMemoryUpdated(ctx, env.Env.OpenAIAssistantVectorStoreID, bot, event)
		}

		return nil
	})

	openaidomain.Init(&openAIClient, env.Env.OpenAIAssistantVectorStoreID)

	commands := []discord.Command{
		NewDJCommand(playerDomain),
	}
	discord.RegisterCommands(bot, env.Env.GuildId, commands...)
	componentInteractionHandlers := []discord.ComponentInteractionHandler{
		chat.NewForgetComponentHandler(&openAIClient),
		player.NewComponentHandler(channelPlayerManager),
	}
	bot.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionMessageComponent {
			discord.HandleComponentInteraction(componentInteractionHandlers, bot, i)

			return
		}

		discord.HandleInteraction(bot, commands, i)
	})
	bot.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Info("connected")
	})
	bot.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		chat.HandleMessageCreate(bot, chatManager, m)
	})

	err = scheduler.Init(bot)
	if err != nil {
		log.Fatal("failed to init scheduler", zap.Error(err))
	}
	err = app.Run(":3000")
	if err != nil {
		log.Fatal("failed to start server", zap.Error(err))
	}

	server.CreateRouter(app, metadata.GetVersion())
}
