package discord

import (
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

func HandleInteraction(bot *Bot, commands []Command, interaction *discordgo.InteractionCreate) {
	log.Info("handling interaction", zap.Any("interaction", interaction))

	for _, command := range commands {
		go command.Handle(bot, interaction)
	}
}
