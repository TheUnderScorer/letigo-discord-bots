package discord

import (
	"app/env"
	"app/errors"
	"app/messages"
	goerrors "errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
)

func prepareErrorReportEmbed(err error) []*discordgo.MessageEmbed {
	var publicError *errors.ErrPublic
	fields := make(map[string]string)
	fields["Error"] = err.Error()

	if goerrors.As(err, &publicError) {
		if publicError.Cause != nil {
			publicError.AddContext("cause", publicError.Cause.Error())
		}

		for key, value := range publicError.Context {
			fields[key] = fmt.Sprintf("%v", value)
		}

	}

	var embedFields []*discordgo.MessageEmbedField
	for key, value := range fields {
		embedFields = append(embedFields, &discordgo.MessageEmbedField{
			Name:  key,
			Value: value,
		})
	}

	return []*discordgo.MessageEmbed{
		{
			Title:  "Error",
			Fields: embedFields,
			Color:  0xff0000,
		},
	}
}

func ReportErrorInteraction(session *discordgo.Session, interaction *discordgo.InteractionCreate, err error) {
	embeds := prepareErrorReportEmbed(err)

	FollowupInteractionMessageAndForget(session, interaction.Interaction, &InteractionReply{
		Content:   messages.Messages.UnknownError,
		Embeds:    embeds,
		Ephemeral: true,
	})
}

func ReportErrorChannel(session *discordgo.Session, cid string, err error) {
	embeds := prepareErrorReportEmbed(err)

	SendMessageComplexAndForget(session, cid, &discordgo.MessageSend{
		Content: messages.Messages.UnknownError,
		Embeds:  embeds,
	})
}

func ReportError(session *discordgo.Session, err error) {
	ReportErrorChannel(session, env.Env.DailyReportChannelId, err)
}
