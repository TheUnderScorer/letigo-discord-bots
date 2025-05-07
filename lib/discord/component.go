package discord

import (
	"context"
	"github.com/bwmarrin/discordgo"
)

type ComponentInteractionHandler interface {
	// Handle processes a Discord component interaction and provides an appropriate response or action via the session.
	Handle(ctx context.Context, interaction *discordgo.InteractionCreate, bot *Bot) error

	// ShouldHandle determines whether the given interaction should be handled by this specific component handler.
	ShouldHandle(interaction *discordgo.InteractionCreate) bool
}

func HandleComponentInteraction(handlers []ComponentInteractionHandler, bot *Bot, interaction *discordgo.InteractionCreate) {
	ctx, cancel := NewInteractionContext(context.Background())
	defer cancel()

	bot.RespondToInteractionAndForget(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})

	for _, handler := range handlers {
		if handler.ShouldHandle(interaction) {
			err := handler.Handle(ctx, interaction, bot)
			if err != nil {
				bot.ReportErrorInteraction(interaction, err)
			}
		}
	}
}
