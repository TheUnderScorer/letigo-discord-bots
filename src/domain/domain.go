package domain

import (
	"app/bots"
	"app/domain/chat"
	"app/domain/interaction"
	"app/llm"
	"app/logging"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

var logger = logging.Get().Named("domain")

type Container struct {
	*interaction.CommandsContainer
	Bots        []*bots.Bot
	ChatManager *chat.Manager
	LlmApi      *llm.API
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
			InitWojciechBot(bot.Session, container.ChatManager, container.LlmApi)
		}
	}

}

func InitWojciechBot(session *discordgo.Session, manager *chat.Manager, llmApi *llm.API) {
	session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		chat.HandleMessageCreate(manager, llmApi, m)
	})
}
