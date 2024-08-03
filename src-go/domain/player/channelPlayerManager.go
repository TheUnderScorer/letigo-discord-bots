package player

import (
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"src-go/logging"
	"sync"
)

type ChannelPlayerManager struct {
	mu      sync.Mutex
	players map[string]*ChannelPlayer
}

func NewChannelPlayerManager() *ChannelPlayerManager {
	return &ChannelPlayerManager{
		players: make(map[string]*ChannelPlayer),
	}
}

func (m *ChannelPlayerManager) GetOrCreate(session *discordgo.Session, channelID string) (*ChannelPlayer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	logger := logging.Get().Named("channelPlayerManager")

	player, ok := m.players[channelID]

	if !ok {
		player, err := NewChannelPlayer(session, channelID, func() {
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
