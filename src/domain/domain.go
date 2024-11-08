package domain

import (
	"app/bots"
	"app/domain/interaction"
	"app/logging"
	"context"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

var logger = logging.Get().Named("domain")

func Init(ctx context.Context) {
	for _, bot := range bots.GetAllFromContext(ctx) {
		session := bot.Session

		session.AddHandler(func(s *discordgo.Session, m *discordgo.InteractionCreate) {
			go interaction.Handle(s, bot.Name, m, ctx)
		})

		session.AddHandler(func(s *discordgo.Session, r *discordgo.GuildMembersChunk) {
			logger.Info("got members chunk", zap.Any("r", r))
		})

		session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
			logger.Info("connected")
		})
	}

}
