package llm

type ChatRole string

const ChatRoleUser = ChatRole("user")
const ChatRoleSystem = ChatRole("system")
const ChatRoleAssistant = ChatRole("assistant")

type ChatMessage struct {
	Contents string   `json:"contents"`
	Role     ChatRole `json:"role"`
}

type Chat struct {
	Messages []*ChatMessage `json:"messages"`
}

func NewChat() *Chat {
	return &Chat{
		Messages: []*ChatMessage{},
	}
}

func (c *Chat) AddMessage(message *ChatMessage) {
	c.Messages = append(c.Messages, message)
}
