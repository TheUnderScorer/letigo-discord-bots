package discord

import (
	"context"
	"github.com/bwmarrin/discordgo"
)

type ComponentInteractionHandler interface {
	// Handle processes a Discord component interaction and provides an appropriate response or action via the session.
	Handle(ctx context.Context, interaction *discordgo.InteractionCreate, session *discordgo.Session) error

	// ShouldHandle determines whether the given interaction should be handled by this specific component handler.
	ShouldHandle(interaction *discordgo.InteractionCreate) bool
}
