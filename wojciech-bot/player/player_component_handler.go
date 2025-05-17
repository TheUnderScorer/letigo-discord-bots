package player

import (
	"app/messages"
	"context"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"lib/discord"
	"lib/errors"
	"lib/logging"
	"lib/util/arrayutil"
)

type ComponentHandler struct {
	channelPlayerManager *ChannelPlayerManager
}

func NewComponentHandler(channelPlayerManager *ChannelPlayerManager) *ComponentHandler {
	return &ComponentHandler{
		channelPlayerManager: channelPlayerManager,
	}
}

func (c ComponentHandler) Handle(_ context.Context, interaction *discordgo.InteractionCreate, bot *discord.Bot) error {
	log := logging.Get().Named("ComponentHandler")

	player, err := c.channelPlayerManager.GetOrCreate(bot, interaction.ChannelID)
	if err != nil {
		return err
	}

	id := interaction.MessageComponentData().CustomID

	log.Debug("handling interaction", zap.String("id", id))

	switch id {
	case string(ButtonPlay):
		return player.Play()

	case string(ButtonPause):
		return player.Pause()

	case string(ButtonNext):
		return player.Next()

	default:
		return errors.NewErrPublic(messages.Messages.UnknownError)
	}
}

func (c ComponentHandler) ShouldHandle(interaction *discordgo.InteractionCreate) bool {
	return arrayutil.Includes(buttons, interaction.MessageComponentData().CustomID)
}
