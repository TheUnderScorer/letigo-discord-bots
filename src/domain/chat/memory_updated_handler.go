package chat

import (
	"app/domain/openai"
	"app/errors"
	"app/messages"
	"app/util/arrayutil"
	"bytes"
	"context"
	"encoding/json"
	"github.com/bwmarrin/discordgo"
)

// HandleMemoryUpdated processes a memory update event and sends a Discord message with associated memory payload data.
func HandleMemoryUpdated(ctx context.Context, vectorStoreID string, session *discordgo.Session, event openai.MemoryUpdated) error {
	payload := ForgetButtonPayload{
		VectorStoreID: vectorStoreID,
		VectorFileID:  event.VectorFileID,
		Content:       event.Content,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "failed to marshal ForgetButtonPayload")
	}

	_, err = session.ChannelMessageSendComplex(event.DiscordThreadID, &discordgo.MessageSend{
		Content: arrayutil.RandomElement(messages.Messages.Chat.NewMemory),
		File: &discordgo.File{
			Name:        "metadata.json",
			ContentType: "application/json",
			Reader:      bytes.NewReader(payloadBytes),
		},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Style:    discordgo.DangerButton,
						Label:    messages.Messages.Chat.ButtonLabelForget,
						CustomID: ForgetButtonID,
						Disabled: false,
					},
				},
			},
		},
	}, discordgo.WithContext(ctx))
	if err != nil {
		return errors.Wrap(err, "failed to send message for ForgetButtonPayload")
	}

	return nil
}
