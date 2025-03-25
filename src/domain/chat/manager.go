package chat

import "sync"

type Manager struct {
	mu    sync.Mutex
	chats map[string]*Chat
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) HasChat(cid string) bool {
	_, ok := m.chats[cid]
	return ok
}
