package chat

import (
	chatevents "app/domain/chat/events"
	"app/errors"
	"app/events"
	"app/llm"
	"app/llm/prompts"
	"app/logging"
	"app/util/arrayutil"
	"context"
	_ "embed"
	goerrors "errors"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"sync"
	"time"
)

const ArchiveDurationMinutes = 60
const MessagesLimit = 100

type DiscordChat struct {
	mu sync.Mutex
	// session stores the Discord session used for managing and interacting with Discord API functionalities.
	session *discordgo.Session
	// parentCid is the thread ID from which the first message originated, and to which thread belongs
	parentCid string
	// log is the logger instance used for structured logging and debugging throughout the DiscordChat lifecycle.
	log *zap.Logger
	// thread in which chat takes place
	thread *discordgo.Channel
	// llmContainer provides access to the LLM (Large Language Model) API for handling chat-related operations in the DiscordChat struct.
	llmContainer *llm.Container
	// firstMessage contains content of the first message that started the thread
	firstMessage *discordgo.Message
	// isFinished indicates if the chat discussion is finished
	isFinished bool
	// TODO When message is deleted, delete it from here
	// chat is the underlying chat used by llm
	chat *llm.Chat
	// onDiscussionEnded is called after discussion is ended
	onDiscussionEnded *func(chat *DiscordChat)
}

func NewDiscordChat(session *discordgo.Session, cid string, llmContainer *llm.Container) *DiscordChat {
	logger := logging.Get().Named("chat").With(zap.String("parentCid", cid), zap.String("session", session.State.User.Username))

	return &DiscordChat{
		session:      session,
		parentCid:    cid,
		log:          logger,
		llmContainer: llmContainer,
		chat:         llm.NewChat(),
	}
}

func (c *DiscordChat) HandleNewMessage(message *discordgo.MessageCreate) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	log := c.log.With(zap.String("messageID", message.ID))
	log.Info("handle new message", zap.String("ID", message.ID))

	if c.isFinished {
		log.Info("already finished")

		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	err := c.ensureThread(ctx, message)
	if err != nil {
		return err
	}

	chat := c.chat

	err = c.session.ChannelTyping(c.thread.ID, discordgo.WithContext(ctx))
	if err != nil {
		log.Error("failed to start typing in channel", zap.Error(err))
	}

	chatMessage := llm.NewDiscordChatMessage(message.Message)
	chat.AddMessages(chatMessage)

	chat, newMessage, newMessageMetadata, err := c.llmContainer.AssistantAPI.Chat(ctx, chat)
	if err != nil {
		log.Error("failed to get new chat from llmContainer", zap.Error(err))

		var tooLongError llm.PromptTooLongError
		if goerrors.As(err, &tooLongError) {
			// DiscordChat got too long for llm to handle, finish the discussion
			return c.EndDiscussion(ctx, message.Message)
		}

		return errors.Wrap(err, "failed to get new chat from llmContainer")
	}
	// Keep the updated chat
	c.chat = chat

	sentMessage, err := c.session.ChannelMessageSend(c.thread.ID, newMessage.Contents, discordgo.WithContext(ctx))
	if err != nil {
		log.Error("failed to send new message", zap.Error(err))
		return errors.Wrap(err, "failed to send new message")
	}
	newMessage.ID = sentMessage.ID

	if newMessageMetadata.IsGoodbye {
		log.Info("bot said goodbye, ending discussion")

		return c.EndDiscussion(ctx, message.Message)
	}

	if newMessageMetadata.IsWorthRemembering {
		go c.remember(sentMessage.ID)
	}

	return nil
}

// EndDiscussion ends the current DiscordChat discussion by reacting to a specified message and marking it as finished.
func (c *DiscordChat) EndDiscussion(ctx context.Context, message *discordgo.Message) error {
	if c.thread == nil {
		return goerrors.New("unable to end discussion, no thread exists")
	}

	err := c.session.MessageReactionAdd(c.thread.ID, message.ID, "👋", discordgo.WithContext(ctx))
	if err != nil {
		return errors.Wrap(err, "failed to react to goodbye message")
	}

	c.isFinished = true

	if c.onDiscussionEnded != nil {
		(*c.onDiscussionEnded)(c)
	}

	return nil
}

// Remember handles process of asking LLM to remember details from the current discord chat. It is meant to run in a separate go routine.
func (c *DiscordChat) remember(messageID string) {
	log := c.log.With(zap.String("messageID", messageID)).Named("remember")

	if c.thread == nil {
		log.Warn("unable to remember, no thread exists")
		return
	}

	// Process only user messages
	userMessages := arrayutil.Filter(c.chat.Messages, func(message *llm.ChatMessage) bool {
		return message.Role == llm.ChatRoleUser
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	systemPrompt := "You are a helpful assistant that parses messages, and extracts details from them. Extract details from this chat messages someone else should remember, and return ONLY them. Try to answer following questions: WHO? and WHAT? Return single sentence that answers these both questions. If message does not contain anything relevant, or you are enable to answer BOTH of these questions, return an empty string. Ignore informations that YOU already know."

	chat := llm.NewChat()
	chat.AddMessages(llm.NewChatMessage(systemPrompt, llm.ChatRoleSystem))
	chat.AddMessages(userMessages...)

	_, reply, _, err := c.llmContainer.ExpensiveAPI.Chat(ctx, chat)
	if err != nil {
		log.Error("failed to extract details to remember", zap.Error(err))
	}

	if reply.Contents == "" {
		log.Error("message has no contents")
	}

	err = events.Dispatch(ctx, chatevents.MemoryDetailsExtracted{
		Details:         reply.Contents,
		DiscordThreadID: c.thread.ID,
	})
	if err != nil {
		log.Error("failed to dispatch MemoryDetailsExtracted event", zap.Error(err))
	}
}

// ensureThread ensures a message thread is created for the given message. If the thread does not exist, it creates one.
func (c *DiscordChat) ensureThread(ctx context.Context, message *discordgo.MessageCreate) error {
	log := c.log.With(zap.String("messageID", message.ID))

	channel, err := c.session.Channel(message.ChannelID)
	if err != nil {
		log.Error("failed to get channel", zap.Error(err))
		return errors.Wrap(err, "failed to get channel")
	}

	if channel.Type == discordgo.ChannelTypeDM {
		log.Info("channel is a DM")
		c.thread = channel
		c.parentCid = channel.ID

		return nil
	}

	if channel.IsThread() {
		log.Info("channel is a thread")

		c.thread = channel
		c.parentCid = channel.ParentID

		return nil
	}

	// Thread does not exist, create it and assign to `channel`
	if c.thread == nil {
		log.Info("first message, creating thread")

		c.firstMessage = message.Message
		// Add the first message to the chat
		c.chat.AddMessages(llm.NewDiscordChatMessage(message.Message))

		threadSummary, err := prompts.SummarizeDiscordThread(ctx, c.llmContainer.AssistantAPI, message.Content)
		if err != nil {
			log.Error("failed to get thread summary", zap.Error(err))
			return errors.Wrap(err, "failed to summarize this message")
		}

		ch, err := c.session.MessageThreadStart(c.parentCid, message.ID, threadSummary.Reply, ArchiveDurationMinutes, discordgo.WithContext(ctx))
		if err != nil {
			log.Error("failed to start thread", zap.Error(err))
			return errors.Wrap(err, "failed to start thread")
		}

		c.log = c.log.With(zap.String("threadID", ch.ID))
		log = c.log

		log.Info("created thread")
		c.thread = ch

		err = c.addThreadMessagesToChat(ctx)
		if err != nil {
			log.Error("failed to sync messages", zap.Error(err))
			return errors.Wrap(err, "failed to sync messages")
		}
	}

	return nil
}

// addThreadMessagesToChat retrieves messages from a Discord thread, processes them and adds them to the chat instance.
func (c *DiscordChat) addThreadMessagesToChat(ctx context.Context) error {
	log := c.log.With(zap.String("threadID", c.thread.ID))

	channelMessages, err := c.session.ChannelMessages(c.thread.ID, MessagesLimit, "", "", "", discordgo.WithContext(ctx))
	if err != nil {
		log.Error("failed to get thread messages", zap.Error(err))
		return errors.Wrap(err, "failed to get thread messages")
	}
	// Exclude empty messages
	channelMessages = arrayutil.Filter(channelMessages, func(message *discordgo.Message) bool {
		return message.Content != ""
	})
	// Reverse the slice, since discord returns messages in order Last to First
	channelMessages = arrayutil.ReverseSlice(channelMessages)

	messagesLen := len(channelMessages)

	if messagesLen > 0 {
		for _, m := range channelMessages {
			var role llm.ChatRole

			// Apply Assistant role to channelMessages sent by bot
			if m.Author.ID == c.session.State.User.ID {
				log.Debug("resolved message to assistant", zap.String("ID", m.ID))
				role = llm.ChatRoleAssistant
			} else {
				log.Debug("resolved message to user", zap.String("ID", m.ID))
				role = llm.ChatRoleUser
			}

			chatMessage := llm.NewDiscordChatMessage(m)
			chatMessage.Role = role
			c.chat.AddMessages(chatMessage)
		}

	}
	return nil
}
