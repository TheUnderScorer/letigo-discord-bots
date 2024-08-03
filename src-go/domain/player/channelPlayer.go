package player

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	dca2 "github.com/jonas747/dca/v2"
	yt "github.com/kkdai/youtube/v2"
	"go.uber.org/zap"
	"io"
	"src-go/dca"
	"src-go/discord"
	"src-go/env"
	errors2 "src-go/errors"
	"src-go/logging"
	"src-go/messages"
	"src-go/util"
	"src-go/youtube"
	"strings"
	"sync"
	"time"
)

type ChannelPlayer struct {
	queue           []*Song
	session         *discordgo.Session
	channelID       string
	voiceConnection *discordgo.VoiceConnection
	logger          *zap.Logger
	isSpeaking      bool
	nextSong        chan *Song
	mu              sync.Mutex
	stream          *dca2.EncodeSession
	currentSong     *Song
	cleanupFuns     []func()
	Disposed        chan bool
}

var ChannelPlayerContextKey = "channelPlayer"

var logger = logging.Get().Named("channelPlayer")

var ytClient = yt.Client{}

func NewChannelPlayer(session *discordgo.Session, channelID string, onDisposed func()) (*ChannelPlayer, error) {
	vc, err := session.ChannelVoiceJoin(env.Cfg.GuildId, channelID, false, true)
	if err != nil {
		return nil, err
	}

	player := &ChannelPlayer{
		session:         session,
		channelID:       channelID,
		voiceConnection: vc,
		logger:          logger.With(zap.String("channelID", channelID)),
		queue:           make([]*Song, 0),
		isSpeaking:      false,
		nextSong:        make(chan *Song),
		currentSong:     nil,
		stream:          nil,
		cleanupFuns:     make([]func(), 0),
	}
	player.cleanupFuns = append(player.cleanupFuns, onDisposed)
	go player.initVoiceConnectionListener()

	return player, nil
}

func (p *ChannelPlayer) Dispose() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.voiceConnection != nil {
		err := p.voiceConnection.Disconnect()
		if err != nil {
			p.logger.Error("failed to disconnect from voice", zap.Error(err))
		}

		if p.stream != nil {
			p.stream.Cleanup()
			p.stream = nil
		}

		select {
		case p.Disposed <- true:
			p.logger.Info("send Disposed message to channel")
		}

		for _, fun := range p.cleanupFuns {
			fun()
			p.cleanupFuns = nil
		}

		p.voiceConnection = nil

	}
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

func (p *ChannelPlayer) initVoiceConnectionListener() {
	p.mu.Lock()
	defer p.mu.Unlock()

	cleanup := p.session.AddHandler(func(s *discordgo.Session, r *discordgo.VoiceStateUpdate) {
		if r.Member.User.ID == s.State.User.ID && (r.BeforeUpdate.ChannelID == p.channelID || r.ChannelID == p.channelID) {
			p.logger.Info("got voice state update", zap.Any("r", r))

			if r.ChannelID == "" {
				p.logger.Info("disconnected from channel")
				p.Dispose()

				return
			}

			if r.ChannelID != p.channelID {
				p.logger.Info("joined new channel", zap.String("channelID", r.ChannelID))
				p.Dispose()

				return
			}

			return
		}
	})
	p.cleanupFuns = append(p.cleanupFuns, cleanup)
}

func (p *ChannelPlayer) ensureReadyVoice() error {
	if p.voiceConnection == nil || !p.voiceConnection.Ready {
		p.logger.Info("voice connection is not ready")

		voice, err := p.session.ChannelVoiceJoin(env.Cfg.GuildId, p.channelID, false, true)

		if err != nil {
			p.logger.Error("failed to re-join voice", zap.Error(err))
			return err
		}

		p.mu.Lock()
		p.voiceConnection = voice
		p.mu.Unlock()
		go p.initVoiceConnectionListener()

		for {
			select {
			case <-p.Disposed:
				return nil

			case <-time.After(1 * time.Minute):
				return errors.New("timeout waiting for connection to be ready")

			default:
				if p.voiceConnection.Ready {
					p.logger.Info("voice connection is ready")
					return nil
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
	}

	p.logger.Info("voice connection was already ready")

	return nil
}

func (p *ChannelPlayer) PlaySong(song *Song) error {
	log := p.logger.With(zap.String("song", song.Name)).With(zap.String("videoID", song.VideoID))

	err := p.ensureReadyVoice()
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

	err := p.voiceConnection.Speaking(true)
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

		case <-p.Disposed:
			p.logger.Info("player Disposed, aborting current playback", zap.Any("song", p.currentSong))
			return nil

		default:
			if !p.isSpeaking {
				p.logger.Info("not speaking, pausing current playback", zap.Any("song", p.currentSong))
				return nil
			}

			if !p.voiceConnection.Ready {
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

			p.voiceConnection.OpusSend <- frame
		}
	}

}

func (p *ChannelPlayer) Pause() error {
	p.isSpeaking = false
	return p.voiceConnection.Speaking(false)
}

func (p *ChannelPlayer) Play() error {
	err := p.ensureReadyVoice()
	if err != nil {
		return err
	}

	err = p.voiceConnection.Speaking(true)
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
