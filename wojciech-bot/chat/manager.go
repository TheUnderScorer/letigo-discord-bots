package chat

import (
	"go.uber.org/zap"
	"lib/discord"
	"lib/llm"
	"lib/logging"
	"lib/util/arrayutil"
	"sync"
)

type Manager struct {
	mu           sync.Mutex
	chats        []*DiscordChat
	bot          *discord.Bot
	log          *zap.Logger
	llmContainer *llm.Container
}

func NewManager(bot *discord.Bot, llm *llm.Container) *Manager {
	log := logging.Get().Named("chat").Named("manager").With(zap.String("bot", bot.State.User.Username))

	return &Manager{
		bot:          bot,
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
		m.log.Info("creating new chat", zap.String("parentCid", cid))
		chat = NewDiscordChat(m.bot, cid, m.llmContainer)
		onDiscussionEnd := func(chat *DiscordChat) {
			m.mu.Lock()
			defer m.mu.Unlock()
			m.log.Info("discussion ended, removing chat", zap.String("parentCid", chat.parentCid))
			m.DeleteChat(chat.thread.ID)
		}
		chat.onDiscussionEnded = &onDiscussionEnd
		m.chats = append(m.chats, chat)
	} else {
		m.log.Info("using existing chat", zap.String("parentCid", cid))
	}

	return chat
}
