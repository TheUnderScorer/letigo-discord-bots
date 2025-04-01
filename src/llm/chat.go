package llm

import "app/util/arrayutil"

type ChatRole string

const ChatRoleUser = ChatRole("user")
const ChatRoleSystem = ChatRole("system")
const ChatRoleAssistant = ChatRole("assistant")

type ChatMessage struct {
	ID        string            `json:"id"`
	Contents  string            `json:"contents"`
	Role      ChatRole          `json:"role"`
	Metadata  map[string]string `json:"metadata"`
	IsGoodbye bool              `json:"is_goodbye"`
}

func NewChatMessage(contents string, role ChatRole) *ChatMessage {
	return &ChatMessage{
		Contents:  contents,
		Role:      role,
		Metadata:  make(map[string]string),
		IsGoodbye: false,
	}
}

func NewChatMessageID(ID string, contents string, role ChatRole) *ChatMessage {
	return &ChatMessage{
		ID:        ID,
		Contents:  contents,
		Role:      role,
		Metadata:  make(map[string]string),
		IsGoodbye: false,
	}
}

func (c *ChatMessage) AddMetadata(key string, value string) {
	c.Metadata[key] = value
}

type Chat struct {
	Messages []*ChatMessage    `json:"messages"`
	Metadata map[string]string `json:"metadata"`

	messageIds []string
}

func NewChat() *Chat {
	return &Chat{
		Messages:   []*ChatMessage{},
		Metadata:   make(map[string]string),
		messageIds: make([]string, 0),
	}
}

func (c *Chat) AddMetadata(key string, value string) {
	c.Metadata[key] = value
}

func (c *Chat) AddMessage(message *ChatMessage) {
	if message.ID != "" && arrayutil.Includes(c.messageIds, message.ID) {
		// Avoid duplicate messages
		return
	}

	c.Messages = append(c.Messages, message)
	if message.ID != "" {
		c.messageIds = append(c.messageIds, message.ID)
	}
}
