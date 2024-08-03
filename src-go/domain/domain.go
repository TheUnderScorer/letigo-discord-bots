package domain

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"src-go/domain/interaction"
)

func Init(discord *discordgo.Session, ctx context.Context) {
	discord.AddHandler(func(s *discordgo.Session, m *discordgo.InteractionCreate) {
		interaction.Handle(s, m, ctx)
	})
}
