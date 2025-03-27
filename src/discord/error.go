package discord

import (
	"app/env"
	"app/errors"
	"app/messages"
	"bytes"
	"encoding/json"
	errors2 "errors"
	"github.com/bwmarrin/discordgo"
)

func ReportErrorChannel(session *discordgo.Session, cid string, err error) {
	var publicError *errors.PublicError

	var files []*discordgo.File
	var messageContent string

	if errors2.As(err, &publicError) {
		if publicError.Cause != nil {
			publicError.AddContext("cause", publicError.Cause.Error())
		}

		errorBytes, marshalErr := json.Marshal(publicError)
		if marshalErr == nil {
			reader := bytes.NewReader(errorBytes)
			files = append(files, &discordgo.File{
				Name:        "error.json",
				ContentType: "application/json",
				Reader:      reader,
			},
			)
		}

		messageContent = publicError.Error()
	} else {
		messageContent = messages.Messages.UnknownError
		errorMessageBytes := []byte(err.Error())
		reader := bytes.NewReader(errorMessageBytes)

		files = append(files, &discordgo.File{
			Name:        "error.txt",
			ContentType: "text/plain",
			Reader:      reader,
		})
	}

	if messageContent != "" {
		SendMessageComplexAndForget(session, cid, &discordgo.MessageSend{
			Content: messageContent,
			Files:   files,
		})
	}
}

func ReportError(session *discordgo.Session, err error) {
	ReportErrorChannel(session, env.Env.DailyReportChannelId, err)
}
