package chat

import (
	"app/llm"
	"app/logging"
	"app/util/arrayutil"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"sync"
)

type Manager struct {
	mu           sync.Mutex
	chats        []*DiscordChat
	session      *discordgo.Session
	log          *zap.Logger
	llmContainer *llm.Container
}

func NewManager(session *discordgo.Session, llm *llm.Container) *Manager {
	log := logging.Get().Named("chat").Named("manager").With(zap.String("session", session.State.User.Username))

	return &Manager{
		session:      session,
		log:          log,
		chats:        make([]*DiscordChat, 0),
		llmContainer: llm,
	}
}

func (m *Manager) GetChat(cid string) *DiscordChat {
	// Find chat using their thread
	for _, c := range m.chats {
		if c.thread != nil && c.thread.ID == cid {
			return c
		}
	}

	return nil
}

func (m *Manager) DeleteChat(cid string) {
	for i, c := range m.chats {
		if c.thread != nil && c.thread.ID == cid {
			m.log.Info("deleting chat", zap.String("cid", cid), zap.Int("index", i))
			m.chats = arrayutil.Delete(m.chats, i)
		}
	}
}

func (m *Manager) HasChat(cid string) bool {
	chat := m.GetChat(cid)
	return chat != nil
}

func (m *Manager) GetOrCreateChat(cid string) *DiscordChat {
	m.mu.Lock()
	defer m.mu.Unlock()

	chat := m.GetChat(cid)
	if chat == nil {
		m.log.Debug("creating new chat", zap.String("parentCid", cid))
		chat = NewDiscordChat(m.session, cid, m.llmContainer)
		onDiscussionEnd := func(chat *DiscordChat) {
			m.mu.Lock()
			defer m.mu.Unlock()
			m.log.Info("discussion ended, removing chat", zap.String("parentCid", chat.parentCid))
			m.DeleteChat(chat.thread.ID)
		}
		chat.onDiscussionEnded = &onDiscussionEnd
		m.chats = append(m.chats, chat)
	} else {
		m.log.Debug("using existing chat", zap.String("parentCid", cid))
	}

	return chat
}
