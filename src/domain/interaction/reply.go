package interaction

import (
	"app/discord"
	"app/errors"
	"app/messages"
	errors2 "errors"
	"github.com/bwmarrin/discordgo"
)

func ReplyToError(err error, s *discordgo.Session, i *discordgo.Interaction) {
	if err != nil {
		var userFriendly *errors.ErrPublic
		if errors2.As(err, &userFriendly) {
			discord.FollowupInteractionErrorAndForget(s, i, err)
		} else {
			discord.FollowupInteractionMessageAndForget(s, i, &discord.InteractionReply{
				Content: messages.Messages.UnknownError,
			})
		}
	}
}
