package voice

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"src-go/env"
	"src-go/logging"
	"sync"
	"time"
)

var logger = logging.Get().Named("voiceManager")

type Manager struct {
	mu         sync.Mutex
	channelID  string
	session    *discordgo.Session
	vc         *discordgo.VoiceConnection
	Disposed   chan bool
	logger     *zap.Logger
	cleanupFns []func()
}

func NewManager(session *discordgo.Session, channelID string, onDisposed func()) (*Manager, error) {
	vc, err := session.ChannelVoiceJoin(env.Cfg.GuildId, channelID, false, true)
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
	m.mu.Lock()
	defer m.mu.Unlock()

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
	m.mu.Lock()
	defer m.mu.Unlock()

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

		voice, err := m.session.ChannelVoiceJoin(env.Cfg.GuildId, m.channelID, false, true)

		if err != nil {
			m.logger.Error("failed to re-join voice", zap.Error(err))
			return err
		}

		m.mu.Lock()
		m.vc = voice
		m.mu.Unlock()

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
