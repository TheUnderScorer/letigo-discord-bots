package discord

import (
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"src-go/logging"
)

var log = logging.Get().Named("discord")

func NewClient(token string) *discordgo.Session {
	discord, err := discordgo.New("Bot " + token)

	if err != nil {
		log.Fatal("failed to init discordgo", zap.Error(err))
	}

	err = discord.Open()
	if err != nil {
		log.Fatal("failed to open discord connection", zap.Error(err))
	}

	return discord
}
