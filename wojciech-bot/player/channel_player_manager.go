package player

import (
	"go.uber.org/zap"
	"lib/discord"
	"lib/logging"
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

func (m *ChannelPlayerManager) GetOrCreate(bot *discord.Bot, channelID string) (*ChannelPlayer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	logger := logging.Get().Named("channelPlayerManager")

	player, ok := m.players[channelID]

	if !ok {
		logger.Info("creating new player", zap.String("channelID", channelID))
		player, err := NewChannelPlayer(bot, channelID, func() {
			logger.Info("player disposed, removing reference", zap.String("channelID", channelID))
			delete(m.players, channelID)
		})

		if err == nil {
			m.players[channelID] = player

			return player, nil
		} else {
			return nil, err
		}
	} else {
		logger.Info("using existing player", zap.String("channelID", channelID))
	}

	return player, nil
}
