package voice

import (
	"app/env"
	"app/logging"
	"context"
	"errors"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"sync"
	"time"
)

var logger = logging.Get().Named("voiceManager")

type Manager struct {
	// Lock for voice connection. Should be used before speaking to avoid duplicate speakers
	Lock sync.Mutex
	// Channel that is emitted when voice connection is disposed
	Disposed chan bool

	// Voice channel ID
	channelID string
	session   *discordgo.Session
	vc        *discordgo.VoiceConnection
	logger    *zap.Logger
	// Cleanup functions that are called when voice connection is disposed
	cleanupFns []func()
}

func NewManager(session *discordgo.Session, channelID string, onDisposed func()) (*Manager, error) {
	vc, err := session.ChannelVoiceJoin(env.Env.GuildId, channelID, false, true)
	if err != nil {
		return nil, err
	}

	manager := &Manager{
		vc:        vc,
		session:   session,
		channelID: channelID,
		logger:    logger.With(zap.String("channelID", channelID)),
	}
	manager.cleanupFns = append(manager.cleanupFns, onDisposed)
	go manager.initVoiceConnectionListener()

	return manager, nil
}

func (m *Manager) Dispose() {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	if m.vc != nil {
		err := m.vc.Disconnect()
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

		m.vc = nil
	}
}

func (m *Manager) VoiceConnection() (*discordgo.VoiceConnection, error) {
	err := m.ReadyVoice()
	if err != nil {
		return nil, err
	}

	return m.vc, nil
}

func (m *Manager) initVoiceConnectionListener() {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	cleanup := m.session.AddHandler(func(s *discordgo.Session, r *discordgo.VoiceStateUpdate) {
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

func (m *Manager) ReadyVoice() error {
	if m.vc == nil || !m.vc.Ready {
		m.logger.Info("voice connection is not ready")

		voice, err := m.session.ChannelVoiceJoin(env.Env.GuildId, m.channelID, false, true)

		if err != nil {
			m.logger.Error("failed to re-join voice", zap.Error(err))
			return err
		}

		m.Lock.Lock()
		m.vc = voice
		m.Lock.Unlock()

		for {
			select {
			case <-m.Disposed:
				return nil

			case <-time.After(1 * time.Minute):
				return errors.New("timeout waiting for connection to be ready")

			default:
				if m.vc.Ready {
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
func (m *Manager) getSpeakerContext(ctx context.Context) (context.Context, func()) {
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		select {
		case <-m.Disposed:
			cancel()

			return

		case <-ctx.Done():
			return
		}
	}()

	return ctx, cancel
}

// Speak sends a message to the voice channel that is managed by this manager. It blocks until the speaker finishes speaking.
func (m *Manager) Speak(speaker Speaker) error {
	return m.SpeakContext(context.Background(), speaker)
}

// SpeakContext sends a message to the voice channel that is managed by this manager. It blocks until the speaker finishes speaking.
func (m *Manager) SpeakContext(ctx context.Context, speaker Speaker) error {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	return m.SpeakNonBlockingContext(ctx, speaker)
}

// SpeakNonBlocking is a non-blocking version of Speak
func (m *Manager) SpeakNonBlocking(speaker Speaker) error {
	return m.SpeakNonBlockingContext(context.Background(), speaker)
}

// SpeakNonBlockingContext is a non-blocking version of Speak
func (m *Manager) SpeakNonBlockingContext(ctx context.Context, speaker Speaker) error {
	vc, err := m.VoiceConnection()
	if err != nil {
		return err
	}

	ctx, cancel := m.getSpeakerContext(ctx)
	defer cancel()

	return speaker.Speak(ctx, vc)
}
