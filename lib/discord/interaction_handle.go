package discord

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

func HandleInteraction(bot *Bot, commands []Command, interaction *discordgo.InteractionCreate) {
	log.Info("handling interaction", zap.Any("interaction", interaction))

	ctx, cancel := NewInteractionContext(context.Background())
	defer cancel()

	for _, command := range commands {
		command.Handle(ctx, bot, interaction)
	}
}
