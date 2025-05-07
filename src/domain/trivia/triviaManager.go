package trivia

import (
	"app/domain/tts"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"lib/aws"
	"lib/logging"
	"sync"
)

// Manager manages trivia per voice channel
type Manager struct {
	mu        sync.Mutex
	triviaMap map[string]*Trivia
	tts       *tts.Client
}

func NewManager(tts *tts.Client) *Manager {
	return &Manager{
		triviaMap: make(map[string]*Trivia),
		tts:       tts,
	}
}

func (m *Manager) Get(channelID string) (*Trivia, bool) {
	trivia, ok := m.triviaMap[channelID]

	return trivia, ok
}

func (m *Manager) GetOrCreate(s3 *aws.S3, session *discordgo.Session, channelID string) (*Trivia, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	logger := logging.Get().Named("triviaManager")

	player, ok := m.triviaMap[channelID]

	if !ok {
		player, err := New(s3, session, m.tts, channelID, func() {
			logger.Info("trivia disposed, removing reference", zap.String("channelID", channelID))
			delete(m.triviaMap, channelID)
		})

		if err == nil {
			m.triviaMap[channelID] = player

			return player, nil
		} else {
			return nil, err
		}
	}

	return player, nil
}
