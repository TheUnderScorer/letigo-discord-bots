package chat

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"lib/discord"
	"lib/errors"
	"lib/util/arrayutil"
	"wojciech-bot/messages"
	"wojciech-bot/openai"
)

type MemoryEmbedField string

var (
	MemoryEmbedFieldVectorStoreID = MemoryEmbedField("Vector Store ID")
	MemoryEmbedFieldVectorFileID  = MemoryEmbedField("Vector File ID")
	MemoryEmbedFileID             = MemoryEmbedField("File ID")
)

func (f MemoryEmbedField) String() string {
	return string(f)
}

// HandleMemoryUpdated processes a memory update event and sends a Discord message with associated memory payload data.
func HandleMemoryUpdated(ctx context.Context, vectorStoreID string, bot *discord.Bot, event openai.MemoryUpdated) error {
	session := bot.Session

	_, err := session.ChannelMessageSendComplex(event.DiscordThreadID, &discordgo.MessageSend{
		Content: arrayutil.RandomElement(messages.Messages.Chat.NewMemory),
		Embeds: []*discordgo.MessageEmbed{
			{
				Author: &discordgo.MessageEmbedAuthor{
					Name: bot.Name,
				},
				Description: event.Content,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  MemoryEmbedFieldVectorStoreID.String(),
						Value: vectorStoreID,
					},
					{
						Name:  MemoryEmbedFieldVectorFileID.String(),
						Value: event.VectorFileID,
					},
					{
						Name:  MemoryEmbedFileID.String(),
						Value: event.FileID,
					},
				},
			},
		},
		Components: ForgetMessageComponent(false),
	}, discordgo.WithContext(ctx))
	if err != nil {
		return errors.Wrap(err, "failed to send message for ForgetButtonPayload")
	}

	return nil
}

func ForgetMessageComponent(disabled bool) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Style:    discordgo.DangerButton,
					Label:    messages.Messages.Chat.ButtonLabelForget,
					CustomID: ForgetButtonID,
					Disabled: disabled,
				},
			},
		},
	}
}
