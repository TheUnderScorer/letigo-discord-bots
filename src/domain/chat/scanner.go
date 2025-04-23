package chat

import (
	"app/llm"
	"app/logging"
	"app/util"
	"app/util/arrayutil"
	"context"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"time"
)

const freshMessageDuration = 24 * time.Hour
const messagesLimit = 100
const foundMessagesLimit = 2
const minDuration = 15 * time.Minute
const maxDuration = 120 * time.Minute

type OnMessageFoundFn func(message *discordgo.Message)

// DiscordChannelScanner is responsible for scanning Discord channels for messages that bot can interact with.
type DiscordChannelScanner struct {
	// session is a Discord session used by bot
	session *discordgo.Session
	// onMessageFoundFn is a callback function invoked when a message is found during the channel scanning process.
	onMessageFoundFn OnMessageFoundFn
	// llmApi is an instance of the LLM API used to interact with the large language model for chat and prompt operations.
	llmApi *llm.API
	// stopChan is used to signal the scanner to stop
	stopChan chan struct{}
	// log is the logger instance for the scanner
	log            *zap.Logger
	tickerDuration time.Duration
}

func NewDiscordChannelScanner(session *discordgo.Session, llmApi *llm.API, onMessageFoundFn OnMessageFoundFn) *DiscordChannelScanner {
	return &DiscordChannelScanner{
		onMessageFoundFn: onMessageFoundFn,
		session:          session,
		llmApi:           llmApi,
		stopChan:         make(chan struct{}),
		log:              logging.Get().Named("Chat").Named("DiscordChannelScanner"),
	}
}

// Start begins the scanning routine
func (d *DiscordChannelScanner) Start() {
	d.log.Info("starting Discord channel scanner")

	// Run the scan immediately on start
	d.scanChannels()

	d.log.Info("scan on start finished")

	d.randomiseDuration()

	ticker := time.NewTicker(d.tickerDuration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			d.scanChannels()
			d.log.Info("scan in ticker finished")
			// After scan, set random duration for next one
			d.randomiseDuration()
			ticker.Reset(d.tickerDuration)
		case <-d.stopChan:
			d.log.Info("stopping Discord channel scanner")
			return
		}
	}
}

func (d *DiscordChannelScanner) randomiseDuration() {
	d.tickerDuration = util.RandomDuration(minDuration, maxDuration)
	log.Info("random duration", zap.Float64("durationMinutes", d.tickerDuration.Minutes()))
}

// Stop stops the scanning routine
func (d *DiscordChannelScanner) Stop() {
	close(d.stopChan)
}

// scanChannels scans all channels for messages that are worthy of reply
func (d *DiscordChannelScanner) scanChannels() {
	ctx, cancel := context.WithTimeout(context.Background(), maxDuration)
	defer cancel()

	d.log.Info("scanning channels for messages")

	for _, guild := range d.session.State.Guilds {
		channels, err := d.session.GuildChannels(guild.ID, discordgo.WithContext(ctx))
		if err != nil {
			d.log.Error("failed to get channels for guild", zap.String("guildID", guild.ID), zap.Error(err))
			continue
		}

		// Filter for text channels only
		textChannels := arrayutil.Filter(channels, func(channel *discordgo.Channel) bool {
			return channel.Type == discordgo.ChannelTypeGuildText
		})

		messages := util.ParallelWithValue(textChannels, func(channel *discordgo.Channel) *discordgo.Message {
			return d.scanChannel(ctx, channel)
		}, 10)

		if len(messages) > 0 {
			randomMessages := arrayutil.RandomElements(messages, foundMessagesLimit)
			for _, message := range randomMessages {
				d.onMessageFoundFn(&message)
			}
		}
	}
}

// scanChannel scans a single channel for messages that are worthy of reply
func (d *DiscordChannelScanner) scanChannel(ctx context.Context, channel *discordgo.Channel) *discordgo.Message {
	d.log.Info("scanning channel for messages", zap.String("channelID", channel.ID), zap.String("channelName", channel.Name))

	messages, err := d.session.ChannelMessages(channel.ID, messagesLimit, "", "", "", discordgo.WithContext(ctx))
	if err != nil {
		d.log.Error("failed to get messages for channel", zap.String("channelID", channel.ID), zap.Error(err))
		return nil
	}

	// Filter messages based on criteria
	worthyMessages := d.filterWorthyMessages(ctx, messages)

	// If we found worthy messages, choose one randomly and pass it to the callback
	if len(worthyMessages) > 0 {
		// Choose a random message
		chosenMessage := arrayutil.RandomElement(worthyMessages)
		d.log.Info("found worthy message", zap.String("messageID", chosenMessage.ID), zap.String("content", chosenMessage.Content))

		return chosenMessage
	}

	return nil
}

// filterWorthyMessages filters messages based on the specified criteria
func (d *DiscordChannelScanner) filterWorthyMessages(ctx context.Context, messages []*discordgo.Message) []*discordgo.Message {
	// Filter messages based on criteria
	worthyMessages := arrayutil.Filter(messages, func(message *discordgo.Message) bool {
		// Ignore messages sent by bots (including us :P)
		if message.Author.Bot {
			return false
		}

		// Ignore messages that have threads
		if message.Thread != nil {
			return false
		}

		// Check if the message is fresh enough
		if time.Since(message.Timestamp) > freshMessageDuration {
			return false
		}

		// Check if the message is long enough and worthy of reply
		return IsWorthyOfReply(ctx, d.llmApi, message)
	})

	return worthyMessages
}
