package trivia

import (
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"src-go/logging"
	"sync"
)

type Manager struct {
	mu      sync.Mutex
	players map[string]*Trivia
}

func NewManager() *Manager {
	return &Manager{
		players: make(map[string]*Trivia),
	}
}

func (m *Manager) GetOrCreate(session *discordgo.Session, channelID string) (*Trivia, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	logger := logging.Get().Named("triviaManager")

	player, ok := m.players[channelID]

	if !ok {
		player, err := New(session, channelID, func() {
			logger.Info("player disposed, removing reference", zap.String("channelID", channelID))
			delete(m.players, channelID)
		})

		if err == nil {
			m.players[channelID] = player

			return player, nil
		} else {
			return nil, err
		}
	}

	return player, nil
}
