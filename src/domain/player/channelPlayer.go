package player

import (
	"app/dca"
	"app/discord"
	"app/domain/voice"
	errors2 "app/errors"
	"app/logging"
	"app/messages"
	"app/util"
	"app/youtube"
	"fmt"
	"github.com/bwmarrin/discordgo"
	dca2 "github.com/jonas747/dca/v2"
	yt "github.com/kkdai/youtube/v2"
	"go.uber.org/zap"
	"io"
	"strings"
	"sync"
	"time"
)

type ChannelPlayer struct {
	queue       []*Song
	session     *discordgo.Session
	channelID   string
	vm          *voice.Manager
	logger      *zap.Logger
	isSpeaking  bool
	nextSong    chan *Song
	mu          sync.Mutex
	stream      *dca2.EncodeSession
	currentSong *Song
}

var ChannelPlayerContextKey = "channelPlayer"

var logger = logging.Get().Named("channelPlayer")

var ytClient = yt.Client{}

func NewChannelPlayer(session *discordgo.Session, channelID string, onDisposed func()) (*ChannelPlayer, error) {
	player := &ChannelPlayer{
		session:     session,
		channelID:   channelID,
		logger:      logger.With(zap.String("channelID", channelID)),
		queue:       make([]*Song, 0),
		isSpeaking:  false,
		nextSong:    make(chan *Song),
		currentSong: nil,
		stream:      nil,
	}
	vm, err := voice.NewManager(session, channelID, func() {
		onDisposed()
		player.Dispose()
	})
	if err != nil {
		return nil, err
	}
	player.vm = vm

	return player, nil
}

func (p *ChannelPlayer) Dispose() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.stream != nil {
		p.stream.Cleanup()
		p.stream = nil
	}

	p.queue = []*Song{}
	p.currentSong = nil

	p.vm.Dispose()
}

func (p *ChannelPlayer) Next() error {
	if len(p.queue) == 0 {
		return errors2.NewUserFriendlyError(messages.Messages.Player.NoMoreSongs)
	}

	p.mu.Lock()
	song := p.queue[0]
	p.queue = p.queue[1:]
	p.mu.Unlock()

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

func (p *ChannelPlayer) PlaySong(song *Song) error {
	log := p.logger.With(zap.String("song", song.Name)).With(zap.String("videoID", song.VideoID))

	err := p.vm.ReadyVoice()
	if err != nil {
		return err
	}

	dcaStream, err := dca.Convert(song.StreamUrl)
	if err != nil {
		log.Error("failed to encode audio", zap.Error(err))
		return err
	}
	p.stream = dcaStream
	p.currentSong = song

	return p.playSession()
}

func (p *ChannelPlayer) playSession() error {
	isFinished := false

	time.Sleep(500 * time.Millisecond)

	defer func() {
		if isFinished {
			p.stream.Cleanup()
			p.stream = nil
		}
	}()

	vc, err := p.vm.VoiceConnection()
	if err != nil {
		return err
	}
	err = vc.Speaking(true)
	if err != nil {
		return err
	}
	p.isSpeaking = true

	discord.SendMessageAndForget(p.session, p.channelID, util.ApplyTokens(util.RandomElement(messages.Messages.Player.NowPlaying), map[string]string{
		"SONG_NAME": p.currentSong.Name,
	}))

	for {
		select {
		case <-p.nextSong:
			p.logger.Info("next song requested, aborting current playback", zap.Any("song", p.currentSong))
			return nil

		case <-p.vm.Disposed:
			p.logger.Info("player Disposed, aborting current playback", zap.Any("song", p.currentSong))
			return nil

		default:
			if !p.isSpeaking {
				p.logger.Info("not speaking, pausing current playback", zap.Any("song", p.currentSong))
				return nil
			}

			if !vc.Ready {
				p.logger.Info("voice connection not ready, pausing current playback", zap.Any("song", p.currentSong))
				return nil
			}

			frame, err := p.stream.OpusFrame()

			if err != nil {
				if err == io.EOF {
					isFinished = true
					p.logger.Info("reached end of stream")

				} else {
					p.logger.Error("failed to get opus frame", zap.Error(err))

				}

				if len(p.queue) > 0 {
					err = p.Next()
					if err != nil {
						p.logger.Error("failed to play next song", zap.Error(err))
					}
				} else {
					err = p.Pause()
					if err != nil {
						p.logger.Error("failed to pause", zap.Error(err))
					}

					go discord.SendMessageAndForget(p.session, p.channelID, messages.Messages.Player.NoMoreSongs)
				}

				return nil
			}

			vc.OpusSend <- frame
		}
	}

}

func (p *ChannelPlayer) Pause() error {
	vc, err := p.vm.VoiceConnection()
	if err != nil {
		return err
	}

	p.isSpeaking = false
	return vc.Speaking(false)
}

func (p *ChannelPlayer) Play() error {
	vc, err := p.vm.VoiceConnection()
	if err != nil {
		return err
	}

	err = vc.Speaking(true)
	if err != nil {
		return err
	}

	if p.stream != nil {
		return p.playSession()
	}

	return nil
}

func (p *ChannelPlayer) Queue() []*Song {
	return p.queue
}

func (p *ChannelPlayer) AddToQueue(url string, userID string) (int, error) {
	info, err := ytClient.GetVideo(url)
	if err != nil {
		return 0, err
	}

	streamUrl, format, err := youtube.GetAudioURL(info.ID)
	if err != nil {
		return 0, err
	}

	song := &Song{
		Url:       url,
		VideoID:   info.ID,
		Name:      info.Title,
		AuthorID:  userID,
		StreamUrl: streamUrl,
		Format:    format,
	}
	itemIndex := len(p.queue)
	p.queue = append(p.queue, song)

	if !p.isSpeaking {
		p.logger.Info("not speaking, playing first queue item", zap.Any("song", song))
		err = p.Next()

		return itemIndex, err
	}

	p.logger.Info("added to queue", zap.Any("song", song))

	return itemIndex, nil
}

func (p *ChannelPlayer) ClearQueue() {
	if len(p.queue) == 0 {
		return
	}

	p.queue = []*Song{}
	p.logger.Info("cleared queue")
}

func (p *ChannelPlayer) ListQueueForDisplay() string {
	items := make([]string, 0)

	for _, song := range p.queue {
		items = append(items, fmt.Sprintf("* %s", append(items, song.Name)))
	}

	return strings.Join(items, "\n")
}
