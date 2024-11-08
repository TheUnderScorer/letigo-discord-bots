package interaction

import (
	"app/discord"
	"app/domain/trivia"
	errors2 "app/errors"
	"app/messages"
	"context"
	"errors"
	"github.com/bwmarrin/discordgo"
	"strings"
)

func HandleComponentInteraction(context context.Context, s *discordgo.Session, cid string, i *discordgo.InteractionCreate) {
	var ufe *errors2.UserFriendlyError

	customID := i.MessageComponentData().CustomID

	if strings.HasPrefix(customID, "trivia") {
		err := trivia.HandleInteraction(context, cid, i)
		if err != nil {
			if errors.As(err, &ufe) {
				go discord.FollowupInteractionErrorAndForget(s, i.Interaction, err)
			} else {
				go discord.FollowupInteractionMessageAndForget(s, i.Interaction, &discord.InteractionReply{
					Content: messages.Messages.UnknownError,
				})
			}
		}

		return
	}
}
