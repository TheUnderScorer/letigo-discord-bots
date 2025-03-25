package chat

import "github.com/bwmarrin/discordgo"

type Chat struct {
	session *discordgo.Session
	cid     string
}

func New(session *discordgo.Session, cid string) *Chat {
	return &Chat{
		session: session,
		cid:     cid,
	}
}
