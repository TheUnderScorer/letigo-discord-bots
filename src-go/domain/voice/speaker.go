package voice

import (
	"context"
	"github.com/bwmarrin/discordgo"
)

type Speaker interface {
	Speak(ctx context.Context, vc *discordgo.VoiceConnection) error
}
