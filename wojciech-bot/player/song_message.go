package player

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"lib/discord"
	"lib/util"
	"lib/util/arrayutil"
	"lib/util/markdownutil"
	"wojciech-bot/messages"
)

type getComponentsFn func() *[]discordgo.MessageComponent

type songMessage struct {
	channelID      string
	playbackState  *playbackState
	discordMessage *discordgo.Message
	bot            *discord.Bot
	getComponents  getComponentsFn
}

// Delete removes the associated discord message if it exists, using the provided context, and resets its reference.
func (m *songMessage) Delete(ctx context.Context) error {
	if m.discordMessage != nil {
		err := m.bot.ChannelMessageDelete(m.channelID, m.discordMessage.ID, discordgo.WithContext(ctx))
		if err != nil {
			return err
		}
		m.discordMessage = nil
	}

	return nil
}

func (m *songMessage) Send(ctx context.Context) error {
	if m.playbackState == nil || m.playbackState.song == nil {
		return m.Delete(ctx)
	}

	song := m.playbackState.song

	msgContent := util.ApplyTokens(arrayutil.RandomElement(messages.Messages.Player.NowPlaying), map[string]string{
		"SONG_NAME": markdownutil.Link(song.Url, song.Name),
	})
	embeds := []*discordgo.MessageEmbed{
		{
			Title: song.Name,
			URL:   song.Url,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: song.ThumbnailUrl,
			},
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Dodane przez",
					Value:  discord.Mention(song.AuthorID),
					Inline: true,
				},
				{
					Value: m.playbackState.String(),
				},
			},
		},
	}
	components := m.getComponents()

	var msg *discordgo.Message
	var err error

	if m.discordMessage != nil {
		msg, err = m.bot.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Embeds:     &embeds,
			ID:         m.discordMessage.ID,
			Channel:    m.discordMessage.ChannelID,
			Components: components,
		}, discordgo.WithContext(ctx))
	} else {
		msg, err = m.bot.ChannelMessageSendComplex(m.channelID, &discordgo.MessageSend{
			Content:    msgContent,
			Embeds:     embeds,
			Components: *components,
		}, discordgo.WithContext(ctx))
	}

	if err != nil {
		return err
	}

	m.discordMessage = msg

	return nil
}
