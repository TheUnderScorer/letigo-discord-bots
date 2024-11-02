package trivia

import (
	"app/domain/tts"
	"app/logging"
	"context"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"sync"
)

// Manager manages trivias per voice channel
type Manager struct {
	mu      sync.Mutex
	trivias map[string]*Trivia
	tts     *tts.Client
}

var ManagerContextKey = "triviaManager"

func NewManager(tts *tts.Client) *Manager {
	return &Manager{
		trivias: make(map[string]*Trivia),
		tts:     tts,
	}
}

func (m *Manager) Get(channelID string) (*Trivia, bool) {
	trivia, ok := m.trivias[channelID]

	return trivia, ok
}

func (m *Manager) GetOrCreate(ctx context.Context, session *discordgo.Session, channelID string) (*Trivia, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	logger := logging.Get().Named("triviaManager")

	player, ok := m.trivias[channelID]

	if !ok {
		player, err := New(ctx, session, m.tts, channelID, func() {
			logger.Info("trivia disposed, removing reference", zap.String("channelID", channelID))
			delete(m.trivias, channelID)
		})

		if err == nil {
			m.trivias[channelID] = player

			return player, nil
		} else {
			return nil, err
		}
	}

	return player, nil
}
