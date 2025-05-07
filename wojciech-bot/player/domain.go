package player

import (
	"context"
	"errors"
	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"go.uber.org/zap"
	"lib/discord"
	errorslib "lib/errors"
	"lib/util"
	"lib/util/arrayutil"
	"strconv"
	"wojciech-bot/messages"
)

type Domain struct {
	playerManager *ChannelPlayerManager
	bot           *discord.Bot
}

func NewDomain(playerManager *ChannelPlayerManager, bot *discord.Bot) *Domain {
	return &Domain{
		playerManager: playerManager,
		bot:           bot,
	}
}

var ErrSongUrlEmpty = errors.New("song url is empty")

func (d *Domain) ensureVoiceChannel(ctx context.Context, interaction *discordgo.Interaction) error {
	channel, err := d.bot.Channel(interaction.ChannelID, discordgo.WithContext(ctx))
	if err != nil {
		log.Error("failed to get channel", zap.Error(err))
		return err
	}

	if channel.Type != discordgo.ChannelTypeGuildVoice {
		log.Error("channel is not a voice channel")

		d.bot.ReplyToInteractionAndForget(interaction, &discord.InteractionReply{
			Content:   messages.Messages.MustBeInVoiceChannel,
			Ephemeral: true,
		})

		return errors.New("must be in a voice channel")
	}

	return nil
}

func (d *Domain) List(ctx context.Context, interaction *discordgo.Interaction) error {
	err := d.ensureVoiceChannel(ctx, interaction)
	if err != nil {
		return err
	}

	channelPlayer, err := d.playerManager.GetOrCreate(d.bot, interaction.ChannelID)
	if err != nil {
		log.Error("failed to get channel player", zap.Error(err))
		return err
	}

	message := channelPlayer.ListQueueForDisplay()
	if message == "" {
		message = messages.Messages.Player.NoMoreSongs
	}
	d.bot.FollowupInteractionMessageAndForget(interaction, &discord.InteractionReply{
		Content: message,
	})

	return nil
}

func (d *Domain) ClearQueue(ctx context.Context, interaction *discordgo.Interaction) error {
	err := d.ensureVoiceChannel(ctx, interaction)
	if err != nil {
		return err
	}

	channelPlayer, err := d.playerManager.GetOrCreate(d.bot, interaction.ChannelID)
	if err != nil {
		log.Error("failed to get channel player", zap.Error(err))
		return err
	}

	channelPlayer.ClearQueue()
	d.bot.DeleteFollowupAndForget(interaction)

	return nil
}

func (d *Domain) Next(ctx context.Context, interaction *discordgo.Interaction) error {
	err := d.ensureVoiceChannel(ctx, interaction)
	if err != nil {
		return err
	}

	channelPlayer, err := d.playerManager.GetOrCreate(d.bot, interaction.ChannelID)
	if err != nil {
		log.Error("failed to get channel player", zap.Error(err))
		return err
	}

	err = channelPlayer.Next()
	if err != nil {
		log.Error("failed to play next", zap.Error(err))
		return err
	}

	d.bot.DeleteFollowupAndForget(interaction)

	return nil
}

func (d *Domain) Play(ctx context.Context, interaction *discordgo.Interaction) error {
	err := d.ensureVoiceChannel(ctx, interaction)
	if err != nil {
		return err
	}

	channelPlayer, err := d.playerManager.GetOrCreate(d.bot, interaction.ChannelID)
	if err != nil {
		log.Error("failed to get channel player", zap.Error(err))
		return err
	}

	err = channelPlayer.Play()
	if err != nil {
		log.Error("failed to play", zap.Error(err))
		return err
	}

	d.bot.DeleteFollowupAndForget(interaction)

	return nil
}

func (d *Domain) Pause(ctx context.Context, interaction *discordgo.Interaction) error {
	err := d.ensureVoiceChannel(ctx, interaction)
	if err != nil {
		return err
	}

	channelPlayer, err := d.playerManager.GetOrCreate(d.bot, interaction.ChannelID)
	if err != nil {
		log.Error("failed to get channel player", zap.Error(err))
		return err
	}

	err = channelPlayer.Pause()
	if err != nil {
		log.Error("failed to pause", zap.Error(err))
		return err
	}

	d.bot.DeleteFollowupAndForget(interaction)

	return nil
}

func (d *Domain) Queue(ctx context.Context, interaction *discordgo.Interaction, songURL string) error {
	if songURL == "" {
		return ErrSongUrlEmpty
	}

	err := d.ensureVoiceChannel(ctx, interaction)
	if err != nil {
		return err
	}

	channelPlayer, err := d.playerManager.GetOrCreate(d.bot, interaction.ChannelID)
	if err != nil {
		log.Error("failed to get channel player", zap.Error(err))
		return err
	}

	order, err := channelPlayer.AddToQueue(songURL, interaction.Member.User.ID)
	if err != nil {
		log.Error("failed to queue song", zap.Error(err))

		d.bot.FollowupInteractionMessageAndForget(interaction, &discord.InteractionReply{
			Content:   messages.Messages.Player.FailedToQueue,
			Ephemeral: true,
		})

		return errorslib.NewErrPublicCause(messages.Messages.Player.FailedToQueue, err)
	}

	var message string
	if order == 0 {
		message = messages.Messages.Player.AddedToQueueAsNext
	} else {
		message = util.ApplyTokens(arrayutil.RandomElement(messages.Messages.Player.AddedToQueue), map[string]string{
			"INDEX": strconv.Itoa(order),
		})
	}

	d.bot.FollowupInteractionMessageAndForget(interaction, &discord.InteractionReply{
		Content: message,
	})

	return nil
}

func (d *Domain) Player(ctx context.Context, interaction *discordgo.Interaction) error {
	channelPlayer, err := d.playerManager.GetOrCreate(d.bot, interaction.ChannelID)
	if err != nil {
		log.Error("failed to get channel player", zap.Error(err))
		return err
	}

	component, err := GetPlayerComponent(channelPlayer)
	if err != nil {
		log.Error("failed to get player component", zap.Error(err))
		d.bot.FollowupInteractionMessageAndForget(interaction, &discord.InteractionReply{
			Content: messages.Messages.UnknownError,
		})
		return err
	}

	d.bot.SendMessageComplexAndForget(interaction.ChannelID, &discordgo.MessageSend{
		Components: *component,
	})
	d.bot.DeleteFollowupAndForget(interaction)
	return nil
}
