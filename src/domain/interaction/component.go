package interaction

import (
	"app/discord"
	"app/domain/trivia"
	errors2 "app/errors"
	"app/messages"
	"errors"
	"github.com/bwmarrin/discordgo"
	"strings"
)

func HandleComponentInteraction(triviaManager *trivia.Manager, s *discordgo.Session, cid string, i *discordgo.InteractionCreate) {
	var ufe *errors2.PublicError

	customID := i.MessageComponentData().CustomID

	if strings.HasPrefix(customID, "trivia") {
		err := trivia.HandleInteraction(triviaManager, cid, i)
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
