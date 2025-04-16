package interaction

import (
	"app/discord"
	"context"
	"github.com/bwmarrin/discordgo"
	"time"
)

func HandleComponentInteraction(handlers []discord.ComponentInteractionHandler, session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	discord.RespondToInteractionAndForget(session, interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})

	for _, handler := range handlers {
		if handler.ShouldHandle(interaction) {
			err := handler.Handle(ctx, interaction, session)
			if err != nil {
				discord.ReportErrorInteraction(session, interaction, err)
			}
		}
	}
}
