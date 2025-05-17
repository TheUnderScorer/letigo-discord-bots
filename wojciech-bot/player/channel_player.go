package player

import (
	"bytes"
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	jonasdca "github.com/jonas747/dca/v2"
	"go.uber.org/zap"
	"io"
	libdiscord "lib/discord"
	"lib/errors"
	"lib/logging"
	"lib/progress"
	"lib/util/markdownutil"
	ytdlp "lib/yt-dlp"
	"math"
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

	stream *jonasdca.EncodeSession
	voice  *libdiscord.Voice
	buffer *bytes.Buffer

	queue       *SongQueue
	currentSong *Song

	mu sync.Mutex

	nextSong       chan *Song
	pauseRequested chan bool

	playbackState *playbackState

	songMessage *songMessage
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

// Dispose releases resources and clears playbackState queue
func (p *ChannelPlayer) Dispose() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.cleanupStream()

	p.queue.Clear()
	p.currentSong = nil

	p.voiceManager.Dispose()
}

// Next advances to the next playbackState in the queue, playing it if available, or signaling the end of the queue if empty.
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
			p.logger.Info("next playbackState dispatched", zap.Any("playbackState", song))
		default:
			p.logger.Info("next playbackState not dispatched", zap.Any("playbackState", song))
		}

		err := p.PlaySong(song)
		if err != nil {
			p.logger.Error("failed to play playbackState", zap.Error(err))
		}
	}()

	return nil
}

// PlaySong plays the provided playbackState by downloading, encoding, and preparing the audio stream for playback.
// Returns an error if voice readiness, opus download, or DCA encoding fails.
func (p *ChannelPlayer) PlaySong(song *Song) error {
	logger := p.logger.With(zap.String("playbackState", song.Name))

	opusBytes, err := ytdlp.DownloadOpus(context.TODO(), song.Url)
	if err != nil {
		logger.Error("failed to download opus", zap.Error(err))
		return err
	}

	buffer := bytes.NewBuffer(opusBytes)
	dcaStream, err := jonasdca.EncodeMem(buffer, jonasdca.StdEncodeOptions)
	if err != nil {
		logger.Error("failed to encode audio", zap.Error(err))
		return err
	}
	logger.Info("prepared dca stream")

	p.mu.Lock()
	defer p.mu.Unlock()

	p.buffer = buffer
	p.voice = libdiscord.NewVoice(dcaStream)

	p.stream = dcaStream
	p.currentSong = song
	p.playbackState = &playbackState{
		song: song,
		isPlaying: func() bool {
			return p.voiceManager.IsSpeaking()
		},
		progress: progress.NewBar(100),
	}

	p.doPlayRoutine()

	return nil
}

func (p *ChannelPlayer) cleanupStream() {
	if p.stream != nil {
		p.stream.Cleanup()
		p.stream = nil
	}
	p.voice = nil
	p.buffer = nil
	p.playbackState = nil
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

// doPlay handles the playback of the current playbackState in the voice channel, managing states and transitions.
func (p *ChannelPlayer) doPlay() error {
	p.mu.Lock()
	defer p.mu.Unlock()

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

	ctx, cancel := p.playbackContext()
	defer cancel()

	// Cleanup previous playbackState message
	if p.songMessage != nil {
		if p.songMessage.playbackState.song != p.currentSong {
			err = p.songMessage.Delete(ctx)
			if err != nil {
				log.Error("failed to dispose playbackState message", zap.Error(err))
			} else {
				log.Debug("removed old song message")
			}

			p.songMessage = nil
		}
	}

	if p.songMessage == nil {
		p.songMessage = &songMessage{
			playbackState: p.playbackState,
			bot:           p.bot,
			channelID:     p.channelID,
			getComponents: func() *[]discordgo.MessageComponent {
				components, _ := GetPlayerComponent(p)
				return components
			},
			discordMessage: nil,
		}
	}

	err = p.songMessage.Send(ctx)
	if err != nil {
		log.Error("failed to send playbackState message", zap.Error(err))
	}

	framesCh := make(chan int, 1)
	defer close(framesCh)
	p.voice.SetFramesSentCh(framesCh)

	go p.trackPlayback(ctx, framesCh)

	err = p.voiceManager.SpeakVoiceContext(ctx, p.voice)
	log.Info("finished playing", zap.Error(err))

	if err != nil {
		// On end of stream, continue playback with next playback State
		if err == io.EOF {
			p.cleanupStream()

			if p.queue.Length() > 0 {
				err = p.Next()
				if err != nil {
					p.logger.Error("failed to play next playbackState", zap.Error(err))
				}
			} else {
				err = p.Pause()
				if err != nil {
					p.logger.Error("failed to pause", zap.Error(err))
				}

				p.bot.SendMessageAndForget(p.channelID, messages.Messages.Player.NoMoreSongs)
			}
		} else {
			log.Error("SpeakVoiceContext returned error", zap.Error(err))

			err := p.songMessage.Send(context.Background())
			if err != nil {
				log.Error("failed to send song message after finishing playback", zap.Error(err))
			}

			return err
		}
	}

	return nil
}

// playbackContext creates a new cancelable context for playback management and triggers cancellation handling in a goroutine.
func (p *ChannelPlayer) playbackContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	go p.handlePlaybackCancellation(ctx, cancel)

	return ctx, cancel
}

// handlePlaybackCancellation listens for playback cancellation signals and cancels the current playback when triggered.
func (p *ChannelPlayer) handlePlaybackCancellation(ctx context.Context, cancel context.CancelFunc) {
	for {
		select {
		case <-p.nextSong:
			p.logger.Info("next playbackState requested, aborting current playback", zap.Any("playbackState", p.currentSong))
			cancel()
			return

		case <-p.pauseRequested:
			p.logger.Info("pause requested, aborting current playback", zap.Any("playbackState", p.currentSong))
			cancel()
			return

		case <-ctx.Done():
			p.logger.Info("context done", zap.Error(ctx.Err()))
			return
		}
	}
}

// trackPlayback tracks the playback progress by periodically updating the elapsed and remaining song duration in real-time.
func (p *ChannelPlayer) trackPlayback(ctx context.Context, framesCh chan int) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var latestFramesSent int
	var hasUpdate bool

	for {
		select {
		case <-ctx.Done():
			p.logger.Debug("trackPlayback is done", zap.Error(ctx.Err()))
			return

		case framesSent := <-framesCh:
			latestFramesSent = framesSent
			hasUpdate = true

		case <-ticker.C:
			if hasUpdate {
				frameDuration, elapsedDuration, percentagePlayed := p.getLatestPlaybackState(latestFramesSent)

				err := p.playbackState.updateProgressPlayed(int64(percentagePlayed))
				if err != nil {
					log.Error("failed to update progress", zap.Error(err))
					continue
				}

				p.playbackState.updateDuration(p.currentSong.Duration - elapsedDuration)

				p.logger.Debug("percentage played",
					zap.Float64("progressPlayed", percentagePlayed),
					zap.Int("framesSent", latestFramesSent),
					zap.Duration("songDuration", p.currentSong.Duration),
					zap.Duration("frameDuration", frameDuration),
				)

				// Re-send message with updated playback state
				err = p.songMessage.Send(ctx)
				if err != nil {
					log.Error("failed to update song message", zap.Error(err))
				}

				hasUpdate = false
			}
		}
	}
}

// getLatestPlaybackState calculates playback metrics such as frame duration, elapsed duration, and percentage played.
func (p *ChannelPlayer) getLatestPlaybackState(latestFramesSent int) (frameDuration time.Duration, elapsedDuration time.Duration, percentagePlayed float64) {
	frameDurationOpt := p.stream.Options().FrameDuration
	frameDurationMs := float64(frameDurationOpt)
	frameDuration = time.Millisecond * time.Duration(frameDurationOpt)

	elapsedMs := float64(latestFramesSent) * frameDurationMs
	elapsedDuration = time.Duration(elapsedMs) * time.Millisecond

	percentagePlayed = (elapsedMs / float64(p.currentSong.Duration.Milliseconds())) * 100
	percentagePlayed = math.Min(percentagePlayed, 100)

	return frameDuration, elapsedDuration, percentagePlayed
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

// AddToQueue adds a playbackState to the queue using the provided URL and user ID, returning the playbackState's index or an error.
func (p *ChannelPlayer) AddToQueue(url string, userID string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	p.logger.Info("adding to queue", zap.String("url", url))
	metadata, err := ytdlp.GetMetadata(ctx, url)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get metadata")
	}

	song := &Song{
		Url:          url,
		Name:         metadata.Title,
		Duration:     metadata.Duration,
		AuthorID:     userID,
		ThumbnailUrl: metadata.ThumbnailUrl,
	}

	itemIndex := p.queue.Length()
	p.queue.Enqueue(song)

	if !p.voiceManager.IsSpeaking() {
		p.logger.Info("not speaking, playing first queue item", zap.Any("playbackState", song))
		err = p.Next()
		return itemIndex, err
	}

	p.logger.Info("added to queue", zap.Any("playbackState", song))
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

// ListQueueForDisplay returns a formatted string representation of the current playbackState queue for display purposes.
func (p *ChannelPlayer) ListQueueForDisplay() string {
	items := make([]string, 0)
	for i, song := range p.queue.List() {
		displayIndex := i + 1
		mention := libdiscord.Mention(song.AuthorID)
		text := fmt.Sprintf("%d. %s (dodane przez %s)", displayIndex, markdownutil.Link(song.Url, song.Name), mention)
		items = append(items, text)
	}
	return strings.Join(items, "\n")
}
