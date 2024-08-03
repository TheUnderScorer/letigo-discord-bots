package trivia

import (
	"github.com/bwmarrin/discordgo"
	"src-go/domain/voice"
)

type Trivia struct {
	session    *discordgo.Session
	channelID  string
	onDisposed func()
	vm         *voice.Manager
}

func New(session *discordgo.Session, channelID string, onDisposed func()) (*Trivia, error) {
	vm, err := voice.NewManager(session, channelID, onDisposed)
	if err != nil {
		return nil, err
	}

	trivia := &Trivia{
		session:    session,
		channelID:  channelID,
		onDisposed: onDisposed,
		vm:         vm,
	}

	return trivia, nil
}
