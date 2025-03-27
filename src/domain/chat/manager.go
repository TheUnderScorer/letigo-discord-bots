package chat

import (
	"app/llm"
	"app/logging"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"sync"
)

type Manager struct {
	mu           sync.Mutex
	chats        []*Chat
	session      *discordgo.Session
	log          *zap.Logger
	llmContainer *llm.Container
}

func NewManager(session *discordgo.Session, llm *llm.Container) *Manager {
	log := logging.Get().Named("chat").Named("manager").With(zap.String("session", session.State.User.Username))

	return &Manager{
		session:      session,
		log:          log,
		chats:        make([]*Chat, 0),
		llmContainer: llm,
	}
}

func (m *Manager) GetChat(cid string) *Chat {
	// Find chat using their thread
	for _, c := range m.chats {
		if c.thread != nil && c.thread.ID == cid {
			return c
		}
	}

	return nil
}

func (m *Manager) HasChat(cid string) bool {
	chat := m.GetChat(cid)
	return chat != nil
}

func (m *Manager) GetOrCreateChat(cid string) *Chat {
	m.mu.Lock()
	defer m.mu.Unlock()

	chat := m.GetChat(cid)
	if chat == nil {
		m.log.Debug("creating new chat", zap.String("parentCid", cid))
		chat = New(m.session, cid, m.llmContainer)
		m.chats = append(m.chats, chat)
	} else {
		m.log.Debug("using existing chat", zap.String("parentCid", cid))
	}

	return chat
}
