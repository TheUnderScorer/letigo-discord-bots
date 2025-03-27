package chat

import (
	"app/errors"
	"app/llm"
	"app/logging"
	"app/util"
	"context"
	_ "embed"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"time"
)

//go:embed traits.txt
var traits []byte

const ArchiveDurationMinutes = 60
const MessagesLimit = 10

type Chat struct {
	// session stores the Discord session used for managing and interacting with Discord API functionalities.
	session *discordgo.Session
	// parentCid is the thread ID from which the first message originated, and to which thread belongs
	parentCid string
	// log is the logger instance used for structured logging and debugging throughout the Chat lifecycle.
	log *zap.Logger
	// thread in which chat takes place
	thread *discordgo.Channel
	// llmContainer provides access to the LLM (Large Language Model) API for handling chat-related operations in the Chat struct.
	llmContainer *llm.Container
}

func New(session *discordgo.Session, cid string, llmContainer *llm.Container) *Chat {
	log := logging.Get().Named("chat").With(zap.String("parentCid", cid), zap.String("session", session.State.User.Username))

	return &Chat{
		session:      session,
		parentCid:    cid,
		log:          log,
		llmContainer: llmContainer,
	}
}

func (c *Chat) HandleNewMessage(message *discordgo.MessageCreate) error {
	log := c.log.With(zap.String("messageID", message.ID))
	log.Debug("handle new message", zap.String("ID", message.ID))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	err := c.ensureThread(ctx, message)
	if err != nil {
		return err
	}

	traitsStr := string(traits)
	log.Debug("applying traits", zap.String("traits", traitsStr))

	chat := llm.NewChat()
	chat.AddMessage(&llm.ChatMessage{
		Role:     llm.ChatRoleSystem,
		Contents: traitsStr,
	})

	// If message was send in thread, let's list messages around it, otherwise, it means that most likely there are no messages yet in the thread
	if message.ChannelID == c.thread.ID {
		messages, err := c.session.ChannelMessages(c.thread.ID, MessagesLimit, "", "", "", discordgo.WithContext(ctx))
		if err != nil {
			log.Debug("failed to get thread messages", zap.Error(err))
			return errors.Wrap(err, "failed to get thread messages")
		}

		if len(messages) > 0 {
			util.ReverseSlice(messages)
			for _, m := range messages {
				if m.Content == "" {
					continue
				}

				var role llm.ChatRole

				// Apply Assistant role to messages sent by bot
				if m.Author.ID == c.session.State.User.ID {
					log.Debug("resolved message to assistant", zap.String("ID", message.ID))
					role = llm.ChatRoleAssistant
				} else {
					log.Debug("resolved message to user", zap.String("ID", message.ID))
					role = llm.ChatRoleUser
				}

				chat.AddMessage(&llm.ChatMessage{
					Role:     role,
					Contents: m.Content,
				})
			}

		}
	}

	err = c.session.ChannelTyping(c.thread.ID)
	if err != nil {
		log.Error("failed to start typing in channel", zap.Error(err))
	}

	chat, newMessage, err := c.llmContainer.ExpensiveAPI.Chat(ctx, chat)
	if err != nil {
		log.Error("failed to get new chat from llmContainer", zap.Error(err))
		return errors.Wrap(err, "failed to get new chat from llmContainer")
	}

	_, err = c.session.ChannelMessageSend(c.thread.ID, newMessage.Contents, discordgo.WithContext(ctx))
	if err != nil {
		log.Error("failed to send new message", zap.Error(err))
		return errors.Wrap(err, "failed to send new message")
	}

	return nil
}

// ensureThread ensures a message thread is created for the given message. If the thread does not exist, it creates one.
func (c *Chat) ensureThread(ctx context.Context, message *discordgo.MessageCreate) error {
	log := c.log.With(zap.String("messageID", message.ID))

	channel, err := c.session.Channel(message.ChannelID)
	if err != nil {
		log.Error("failed to get channel", zap.Error(err))
		return errors.Wrap(err, "failed to get channel")
	}

	if channel.IsThread() {
		log.Debug("channel is a thread started by us before")

		c.thread = channel
		c.parentCid = channel.ParentID

		return nil
	}

	if c.thread == nil {
		log.Debug("first message, creating thread")

		ch, err := c.session.MessageThreadStart(c.parentCid, message.ID, "Thread", ArchiveDurationMinutes, discordgo.WithContext(ctx))
		if err != nil {
			log.Debug("failed to start thread", zap.Error(err))
			return errors.Wrap(err, "failed to start thread")
		}

		log.Debug("created thread thread", zap.String("threadID", ch.ID))
		c.thread = ch
	}

	return nil
}
