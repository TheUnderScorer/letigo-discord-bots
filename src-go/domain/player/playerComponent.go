package player

import (
	"errors"
	"github.com/bwmarrin/discordgo"
)

type ButtonID string

var (
	ButtonPlay  = ButtonID("play")
	ButtonPause = ButtonID("pause")
	ButtonNext  = ButtonID("next")
)

func GetPlayerComponent(player *ChannelPlayer) (*[]discordgo.MessageComponent, error) {
	if player == nil {
		return &[]discordgo.MessageComponent{}, errors.New("player is nil")
	}

	var actionBtn discordgo.Button

	if player.isSpeaking {
		actionBtn = discordgo.Button{
			Style:    discordgo.PrimaryButton,
			CustomID: string(ButtonPause),
			Disabled: false,
			Emoji: &discordgo.ComponentEmoji{
				Name: "⏸️",
			},
		}
	} else {
		actionBtn = discordgo.Button{
			Style:    discordgo.PrimaryButton,
			CustomID: string(ButtonPlay),
			Disabled: false,
			Emoji: &discordgo.ComponentEmoji{
				Name: "▶️",
			},
		}
	}

	return &[]discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					actionBtn,
					discordgo.Button{
						Style:    discordgo.SecondaryButton,
						CustomID: string(ButtonNext),
						Disabled: len(player.queue) == 0,
						Emoji: &discordgo.ComponentEmoji{
							Name: "⏭️",
						},
					},
				},
			},
		},
		nil
}
