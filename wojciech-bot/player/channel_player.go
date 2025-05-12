package player

import (
	"bytes"
	"context"
	"github.com/charmbracelet/log"
	jonasdca "github.com/jonas747/dca/v2"
	"go.uber.org/zap"
	"io"
	libdiscord "lib/discord"
	"lib/errors"
	"lib/logging"
	"lib/util"
	"lib/util/arrayutil"
	"lib/util/markdownutil"
	ytdlp "lib/yt-dlp"
	"strings"
	"sync"
	"time"
	"wojciech-bot/env"
	"wojciech-bot/messages"
)

// ChannelPlayer manages audio playback and queue in a Discord voice channel.
// It handles playing, queuing, and streaming of audio tracks using Discord voice capabilities.
type ChannelPlayer struct {
	logger *zap.Logger

	bot          *libdiscord.Bot
	channelID    string
	voiceManager *libdiscord.VoiceManager

	stream      *jonasdca.EncodeSession
	queue       *SongQueue
	currentSong *Song

	mu sync.Mutex

	nextSong       chan *Song
	pauseRequested chan bool
}

var logger = logging.Get().Named("channelPlayer")

// NewChannelPlayer initializes a new ChannelPlayer for managing audio playback in a specific channel.
// It takes a bot instance, a channel ID, and a callback function executed upon disposal.
// Returns a pointer to the created ChannelPlayer and an error if initialization fails.
func NewChannelPlayer(bot *libdiscord.Bot, channelID string, onDisposed func()) (*ChannelPlayer, error) {
	player := &ChannelPlayer{
		bot:            bot,
		channelID:      channelID,
		logger:         logger.With(zap.String("channelID", channelID)),
		queue:          NewSongQueue(),
		nextSong:       make(chan *Song),
		pauseRequested: make(chan bool),
		currentSong:    nil,
		stream:         nil,
	}
	voiceManager, err := libdiscord.NewVoiceManager(bot, env.Env.GuildId, channelID, func() {
		onDisposed()
		player.Dispose()
	})
	if err != nil {
		return nil, err
	}
	player.voiceManager = voiceManager

	return player, nil
}

// Dispose releases resources and clears song queue
func (p *ChannelPlayer) Dispose() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.stream != nil {
		p.stream.Cleanup()
		p.stream = nil
	}

	p.queue.Clear()
	p.currentSong = nil

	p.voiceManager.Dispose()
}

// Next advances to the next song in the queue, playing it if available, or signaling the end of the queue if empty.
func (p *ChannelPlayer) Next() error {
	if p.queue.Length() == 0 {
		return errors.NewErrPublic(messages.Messages.Player.NoMoreSongs)
	}

	song := p.queue.Dequeue()
	if song == nil {
		p.logger.Info("queue is empty")
		return p.Pause()
	}

	go func() {
		select {
		case p.nextSong <- song:
			p.logger.Info("next song dispatched", zap.Any("song", song))
		default:
			p.logger.Info("next song not dispatched", zap.Any("song", song))
		}

		err := p.PlaySong(song)
		if err != nil {
			p.logger.Error("failed to play song", zap.Error(err))
		}
	}()

	return nil
}

// PlaySong plays the provided song by downloading, encoding, and preparing the audio stream for playback.
// Returns an error if voice readiness, opus download, or DCA encoding fails.
func (p *ChannelPlayer) PlaySong(song *Song) error {
	logger := p.logger.With(zap.String("song", song.Name))

	opusBytes, err := ytdlp.DownloadOpus(context.TODO(), song.Url)
	if err != nil {
		logger.Error("failed to download opus", zap.Error(err))
		return err
	}

	dcaStream, err := jonasdca.EncodeMem(bytes.NewBuffer(opusBytes), jonasdca.StdEncodeOptions)
	if err != nil {
		logger.Error("failed to encode audio", zap.Error(err))
		return err
	}
	logger.Info("prepared dca stream")
	p.stream = dcaStream
	p.currentSong = song

	p.doPlayRoutine()

	return nil
}

// doPlayRoutine invokes the doPlay method in a separate goroutine, handling playback errors and logging them.
func (p *ChannelPlayer) doPlayRoutine() {
	go func() {
		err := p.doPlay()
		if err != nil {
			log.Error("failed to play", zap.Error(err))
		}
	}()
}

// doPlay handles the playback of the current song in the voice channel, managing states and transitions.
func (p *ChannelPlayer) doPlay() error {
	err := p.bot.UpdateListeningStatus(p.currentSong.Name)
	if err != nil {
		log.Error("failed to update listening status", zap.Error(err))
	}

	time.Sleep(500 * time.Millisecond)

	defer func() {
		err := p.bot.UpdateListeningStatus("")
		if err != nil {
			log.Error("failed to clear listening status", zap.Error(err))
		}
	}()

	p.bot.SendMessageAndForget(p.channelID, util.ApplyTokens(arrayutil.RandomElement(messages.Messages.Player.NowPlaying), map[string]string{
		"SONG_NAME": p.currentSong.Name,
	}))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			select {
			case <-p.nextSong:
				p.logger.Info("next song requested, aborting current playback", zap.Any("song", p.currentSong))
				cancel()
				return

			case <-p.pauseRequested:
				p.logger.Info("pause requested, aborting current playback", zap.Any("song", p.currentSong))
				cancel()
				return

			case <-ctx.Done():
				p.logger.Info("context done")
				return
			}
		}
	}()

	err = p.voiceManager.SpeakOpusReaderContext(ctx, p.stream)
	log.Info("finished playing", zap.Error(err))
	if err != nil {
		// On end of stream, continue playback with next song
		if err == io.EOF {
			// Cleanup old stream
			if p.stream != nil {
				p.stream.Cleanup()
				p.stream = nil
			}

			if p.queue.Length() > 0 {
				err = p.Next()
				if err != nil {
					p.logger.Error("failed to play next song", zap.Error(err))
				}
			} else {
				err = p.Pause()
				if err != nil {
					p.logger.Error("failed to pause", zap.Error(err))
				}

				p.bot.SendMessageAndForget(p.channelID, messages.Messages.Player.NoMoreSongs)
			}
		} else {
			log.Error("SpeakOpusReaderContext returned error", zap.Error(err))

			return err
		}
	}

	return nil
}

// Pause stops the bot from speaking in the voice channel and updates its speaking state. Returns an error if unsuccessful.
func (p *ChannelPlayer) Pause() error {
	p.logger.Info("pausing")
	select {
	case p.pauseRequested <- true:
		p.logger.Info("pauseRequested sent")
		return nil

	default:
		p.logger.Info("pauseRequested not sent")
		return nil
	}

}

// Play starts playing the current audio stream in the voice channel if available, establishing the voice connection if needed.
func (p *ChannelPlayer) Play() error {
	vc, err := p.voiceManager.VoiceConnection()
	if err != nil {
		return err
	}

	err = vc.Speaking(true)
	if err != nil {
		return err
	}

	if p.stream != nil {
		p.doPlayRoutine()

		return nil
	}

	return nil
}

// Queue returns the current list of songs in the ChannelPlayer's queue.
func (p *ChannelPlayer) Queue() []*Song {
	return p.queue.List()
}

// AddToQueue adds a song to the queue using the provided URL and user ID, returning the song's index or an error.
func (p *ChannelPlayer) AddToQueue(url string, userID string) (int, error) {
	p.logger.Info("adding to queue", zap.String("url", url))
	title, err := ytdlp.GetTitle(context.TODO(), url)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get title")
	}

	song := &Song{
		Url:      url,
		Name:     title,
		AuthorID: userID,
	}

	itemIndex := p.queue.Length()
	p.queue.Enqueue(song)

	if !p.voiceManager.IsSpeaking() {
		p.logger.Info("not speaking, playing first queue item", zap.Any("song", song))
		err = p.Next()
		return itemIndex, err
	}

	p.logger.Info("added to queue", zap.Any("song", song))
	return itemIndex, nil
}

// ClearQueue removes all songs from the queue and logs that the queue has been cleared. Does nothing if empty.
func (p *ChannelPlayer) ClearQueue() {
	if p.queue.Length() == 0 {
		return
	}
	p.queue.Clear()
	p.logger.Info("cleared queue")
}

// ListQueueForDisplay returns a formatted string representation of the current song queue for display purposes.
func (p *ChannelPlayer) ListQueueForDisplay() string {
	items := make([]string, 0)
	for _, song := range p.queue.List() {
		items = append(items, "* "+markdownutil.Link(song.Url, song.Name))
	}
	return strings.Join(items, "\n")
}
