package discord

import (
	"app/logging"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

var log = logging.Get().Named("discord")

func NewClient(token string) *discordgo.Session {
	if token == "" {
		log.Fatal("token is empty")
	}

	discord, err := discordgo.New("Bot " + token)
	discord.SyncEvents = false

	if err != nil {
		log.Fatal("failed to init discordgo", zap.Error(err))
	}
	discord.Identify.Intents = discordgo.IntentsAll
	discord.StateEnabled = true

	err = discord.Open()
	if err != nil {
		log.Fatal("failed to open discord connection", zap.Error(err))
	}
	log.Info("discord connection opened")

	return discord
}
