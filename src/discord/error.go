package discord

import (
	"app/env"
	"app/errors"
	"app/messages"
	goerrors "errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
)

func ReportErrorChannel(session *discordgo.Session, cid string, err error) {
	var publicError *errors.PublicError
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

	SendMessageComplexAndForget(session, cid, &discordgo.MessageSend{
		Content: messages.Messages.UnknownError,
		Embeds: []*discordgo.MessageEmbed{
			{
				Title:  "Error",
				Fields: embedFields,
				Color:  0xff0000,
			},
		},
	})
}

func ReportError(session *discordgo.Session, err error) {
	ReportErrorChannel(session, env.Env.DailyReportChannelId, err)
}
