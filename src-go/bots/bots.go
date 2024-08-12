package bots

import (
	"github.com/bwmarrin/discordgo"
	"src-go/discord"
)

type BotName string

var (
	BotNameWojciech     = BotName("wojciech")
	BotNameTadeuszSznuk = BotName("tadeusz-sznuk")
)

type Bot struct {
	Name    BotName
	token   string
	Session *discordgo.Session
}

func NewBot(name BotName, token string) *Bot {
	return &Bot{
		Name:    name,
		token:   token,
		Session: discord.NewClient(token),
	}
}
