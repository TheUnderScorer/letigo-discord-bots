package domain

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"src-go/domain/interaction"
	"src-go/logging"
)

var logger = logging.Get().Named("domain")

func Init(discord *discordgo.Session, ctx context.Context) {
	discord.AddHandler(func(s *discordgo.Session, m *discordgo.InteractionCreate) {
		go interaction.Handle(s, m, ctx)
	})

	discord.AddHandler(func(s *discordgo.Session, r *discordgo.GuildMembersChunk) {
		logger.Info("got members chunk", zap.Any("r", r))
	})

	discord.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		logger.Info("connected")
	})
}
