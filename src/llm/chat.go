package llm

import (
	"app/discord"
	"app/util/arrayutil"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"sync"
)

type ChatRole string

const ChatRoleUser = ChatRole("user")
const ChatRoleSystem = ChatRole("system")
const ChatRoleAssistant = ChatRole("assistant")

type ChatMessage struct {
	ID         string            `json:"id"`
	Contents   string            `json:"contents"`
	Role       ChatRole          `json:"role"`
	Metadata   map[string]string `json:"metadata"`
	AuthorName string            `json:"author_name"`
}

func NewAssistantChatMessage(contents string, messageID string) *ChatMessage {
	return &ChatMessage{
		ID:         messageID,
		Contents:   contents,
		Role:       ChatRoleAssistant,
		AuthorName: "Assistant",
		Metadata:   make(map[string]string),
	}
}

func NewDiscordChatMessage(message *discordgo.Message) *ChatMessage {
	var authorName string
	friend, ok := discord.GetFriend(message.Author.ID)
	if ok {
		authorName = friend.FirstName
	} else {
		authorName = message.Author.Username
	}

	return &ChatMessage{
		ID:         message.ID,
		Contents:   message.ContentWithMentionsReplaced(),
		Role:       ChatRoleUser,
		AuthorName: authorName,
		Metadata:   make(map[string]string),
	}
}

func NewUserChatMessage(contents string, messageID string, authorName string) *ChatMessage {
	return &ChatMessage{
		ID:         messageID,
		Contents:   contents,
		Role:       ChatRoleUser,
		AuthorName: authorName,
		Metadata:   make(map[string]string),
	}
}

func NewChatMessage(contents string, role ChatRole) *ChatMessage {
	return &ChatMessage{
		Contents: contents,
		Role:     role,
		Metadata: make(map[string]string),
	}
}

func NewChatMessageID(ID string, contents string, role ChatRole) *ChatMessage {
	return &ChatMessage{
		ID:       ID,
		Contents: contents,
		Role:     role,
		Metadata: make(map[string]string),
	}
}

func (c *ChatMessage) AddMetadata(key string, value string) {
	c.Metadata[key] = value
}

func (c *ChatMessage) ChatMessage() string {
	if c.AuthorName == "" {
		return c.Contents
	}

	return fmt.Sprintf("%s: %s", c.AuthorName, c.Contents)
}

// ChatReplyMetadata represents metadata about a response in a chat interaction.
// It indicates if the response has significance worth remembering or signals the end of a conversation.
type ChatReplyMetadata struct {
	IsWorthRemembering bool
	IsGoodbye          bool
}

type Chat struct {
	mu       sync.Mutex
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

func (c *Chat) AddMessages(messages ...*ChatMessage) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, message := range messages {
		if message.ID != "" && arrayutil.Includes(c.messageIds, message.ID) {
			// Avoid duplicate messages
			continue
		}

		c.Messages = append(c.Messages, message)
		if message.ID != "" {
			c.messageIds = append(c.messageIds, message.ID)
		}
	}
}
