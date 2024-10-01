package interaction

import (
	errors2 "errors"
	"github.com/bwmarrin/discordgo"
	"src-go/discord"
	"src-go/errors"
	"src-go/messages"
)

func ReplyToError(err error, s *discordgo.Session, i *discordgo.Interaction) {
	if err != nil {
		var userFriendly *errors.UserFriendlyError
		if errors2.As(err, &userFriendly) {
			discord.FollowupInteractionErrorAndForget(s, i, err)
		} else {
			discord.FollowupInteractionMessageAndForget(s, i, &discord.InteractionReply{
				Content: messages.Messages.UnknownError,
			})
		}
	}
}
