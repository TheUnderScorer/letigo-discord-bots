package discord

import (
	"context"
	"errors"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca/v2"
	"go.uber.org/zap"
	"io"
	"sync"
	"time"
)

type VoiceManager struct {
	// Lock for voice connection. Should be used before speaking to avoid duplicate speakers
	Lock sync.Mutex
	// Channel that is emitted when voice connection is disposed
	Disposed chan bool

	// Voice channel ID
	channelID       string
	guildID         string
	bot             *Bot
	voiceConnection *discordgo.VoiceConnection
	logger          *zap.Logger
	// Cleanup functions that are called when voice connection is disposed
	cleanupFns []func()
}

func NewVoiceManager(bot *Bot, guildID string, channelID string, onDisposed func()) (*VoiceManager, error) {
	vc, err := bot.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return nil, err
	}

	manager := &VoiceManager{
		voiceConnection: vc,
		bot:             bot,
		channelID:       channelID,
		guildID:         guildID,
		logger:          logger.With(zap.String("channelID", channelID)),
	}
	manager.cleanupFns = append(manager.cleanupFns, onDisposed)
	go manager.initVoiceConnectionListener()

	return manager, nil
}

func (m *VoiceManager) Dispose() {
	if m.voiceConnection != nil {
		err := m.voiceConnection.Disconnect()
		if err != nil {
			m.logger.Error("failed to disconnect from voice", zap.Error(err))
		}

		select {
		case m.Disposed <- true:
			m.logger.Info("send Disposed message to channel")

		default:
			m.logger.Warn("failed to send Disposed message to channel")
		}

		for _, fun := range m.cleanupFns {
			fun()
		}

		m.cleanupFns = nil

		m.voiceConnection = nil
	}
}

func (m *VoiceManager) VoiceConnection() (*discordgo.VoiceConnection, error) {
	err := m.ReadyVoice()
	if err != nil {
		return nil, err
	}

	return m.voiceConnection, nil
}

func (m *VoiceManager) initVoiceConnectionListener() {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	cleanup := m.bot.AddHandler(func(s *discordgo.Session, r *discordgo.VoiceStateUpdate) {
		if r.Member.User.ID == s.State.User.ID && (r.BeforeUpdate != nil && r.BeforeUpdate.ChannelID == m.channelID || r.ChannelID == m.channelID) {
			m.logger.Info("got voice state update", zap.Any("r", r))

			if r.ChannelID == "" {
				m.logger.Info("disconnected from channel")
				m.Dispose()

				return
			}

			if r.ChannelID != m.channelID {
				m.logger.Info("joined new channel", zap.String("channelID", r.ChannelID))
				m.Dispose()

				return
			}

			return
		}
	})

	m.cleanupFns = append(m.cleanupFns, cleanup)
}

func (m *VoiceManager) ReadyVoice() error {
	if m.voiceConnection == nil || !m.voiceConnection.Ready {
		m.logger.Info("voice connection is not ready")

		voice, err := m.bot.ChannelVoiceJoin(m.guildID, m.channelID, false, true)

		if err != nil {
			m.logger.Error("failed to re-join voice", zap.Error(err))
			return err
		}

		m.Lock.Lock()
		m.voiceConnection = voice
		m.Lock.Unlock()

		for {
			select {
			case <-m.Disposed:
				return nil

			case <-time.After(1 * time.Minute):
				return errors.New("timeout waiting for connection to be ready")

			default:
				if m.voiceConnection.Ready {
					m.logger.Info("voice connection is ready")
					return nil
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
	}

	m.logger.Info("voice connection was already ready")

	return nil
}

// getSpeakerContext returns a context that is canceled when the manager is disposed
func (m *VoiceManager) getSpeakerContext(ctx context.Context) (context.Context, func()) {
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		for {
			select {
			case <-m.Disposed:
				cancel()

				return

			case <-ctx.Done():
				return
			}
		}
	}()

	return ctx, cancel
}

func (m *VoiceManager) IsSpeaking() bool {
	if m.Lock.TryLock() {
		defer m.Lock.Unlock()
		return false
	}

	return true
}

// Speak sends a message to the voice channel that is managed by this manager. It blocks until the speaker finishes speaking.
func (m *VoiceManager) Speak(speaker Speaker) error {
	return m.SpeakContext(context.Background(), speaker)
}

// SpeakContext sends a message to the voice channel that is managed by this manager. It blocks until the speaker finishes speaking.
func (m *VoiceManager) SpeakContext(ctx context.Context, speaker Speaker) error {
	return m.doSpeakContext(ctx, speaker)
}

// SpeakOpusReaderContext streams audio frames from a dca.OpusReader to the voice connection until the context is done or disposed.
func (m *VoiceManager) SpeakOpusReaderContext(ctx context.Context, reader dca.OpusReader) error {
	m.logger.Debug("starting to speak")

	vc, err := m.VoiceConnection()
	if err != nil {
		m.logger.Error("failed to get voice connection", zap.Error(err))
		return err
	}

	m.Lock.Lock()
	defer m.Lock.Unlock()

	m.logger.Debug("lock acquired, starting to speak")

	err = vc.Speaking(true)
	if err != nil {
		m.logger.Error("failed to send speaking notification", zap.Error(err))
		return err
	}

	defer func() {
		err = vc.Speaking(false)
		if err != nil {
			m.logger.Error("failed to send speaking notification end", zap.Error(err))
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-m.Disposed:
			m.logger.Info("player disposed, aborting playback")
			return errors.New("player disposed")

		default:
			if !vc.Ready {
				m.logger.Info("voice connection is not ready")
				continue
			}

			frame, err := reader.OpusFrame()

			if err != nil {
				if err == io.EOF {
					m.logger.Info("reached end of stream")
				} else {
					m.logger.Error("failed to read opus frame", zap.Error(err))
				}

				return err
			}

			vc.OpusSend <- frame
		}
	}
}

func (m *VoiceManager) doSpeakContext(ctx context.Context, speaker Speaker) error {
	vc, err := m.VoiceConnection()
	if err != nil {
		return err
	}

	m.Lock.Lock()
	defer m.Lock.Unlock()

	ctx, cancel := m.getSpeakerContext(ctx)
	defer cancel()

	return speaker.Speak(ctx, vc)
}
