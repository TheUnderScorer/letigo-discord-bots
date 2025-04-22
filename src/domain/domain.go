package domain

import (
	"app/bots"
	"app/discord"
	"app/domain/chat"
	"app/domain/interaction"
	"app/domain/openai"
	"app/env"
	"app/events"
	"app/llm"
	"app/logging"
	"context"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

var logger = logging.Get().Named("domain")

type Container struct {
	*interaction.CommandsContainer
	Bots         []*bots.Bot
	ChatManager  *chat.Manager
	LlmApi       *llm.API
	LlmContainer *llm.Container
}

func Init(container *Container) {
	for _, bot := range container.Bots {
		session := bot.Session

		session.AddHandler(func(s *discordgo.Session, m *discordgo.InteractionCreate) {
			interaction.Handle(s, bot.Name, m, container.CommandsContainer)
		})

		session.AddHandler(func(s *discordgo.Session, r *discordgo.GuildMembersChunk) {
			logger.Info("got members chunk", zap.Any("r", r))
		})

		session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
			logger.Info("connected")
		})

		if bot.Name == bots.BotNameWojciech {
			InitWojciechBot(bot, container.ChatManager, container.LlmContainer)
		}
	}

}

func InitWojciechBot(bot *bots.Bot, manager *chat.Manager, llmContainer *llm.Container) {
	session := bot.Session
	session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		chat.HandleMessageCreate(session, manager, llmContainer.FreeAPI, m)
	})
	chatScanner := chat.NewDiscordChannelScanner(session, llmContainer.FreeAPI, func(message *discordgo.Message) {
		err := bot.Session.MessageReactionAdd(message.ChannelID, message.ID, discord.ReactionSeen)
		if err != nil {
			logger.Error("failed to add seen reaction", zap.Error(err))
		}

		discordChat := manager.GetOrCreateChat(message.ChannelID)
		err = discordChat.HandleNewMessage(message)
		if err != nil {
			logger.Error("failed to handle new message", zap.Error(err))
		}
	})
	go chatScanner.Start()

	events.Handle(func(ctx context.Context, event openai.MemoryUpdated) error {
		if event.DiscordThreadID != "" {
			return chat.HandleMemoryUpdated(ctx, env.Env.OpenAIAssistantVectorStoreID, bot, event)
		}

		return nil
	})
}
