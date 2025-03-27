package chat

import (
	"app/llm"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

func HandleMessageCreate(session *discordgo.Session, manager *Manager, llmApi *llm.API, m *discordgo.MessageCreate) {
	log := chatLog.With(zap.String("messageID", m.ID))
	log.Debug("handling message create")

	channel, err := session.Channel(m.ChannelID)
	if err != nil {
		log.Error("failed to get message channel", zap.Error(err))
		return
	}

	if m.Author.ID == session.State.User.ID {
		log.Debug("skip this message, because it was sent by us")

		return
	}

	if manager.HasChat(m.ChannelID) {
		log.Debug("already have chat", zap.String("channelID", m.ChannelID), zap.String("messageID", m.ID))

		chat := manager.GetOrCreateChat(m.ChannelID)
		err := chat.HandleNewMessage(m)
		if err != nil {
			log.Error("failed to handle new message in existing chat", zap.Error(err))
		}

		return
	}

	if channel.IsThread() && channel.OwnerID == session.State.User.ID {
		log.Debug("channel is a thread started by us before")
	}

	// Check if this channel is a thread, and if it was started by us before
	if (channel.IsThread() && channel.OwnerID == session.State.User.ID) ||
		// Otherwise, check if message is worthy of reply
		IsWorthyOfReply(llmApi, m.Content) {
		log.Info("message is worthy of reply", zap.String("content", m.Content))

		chat := manager.GetOrCreateChat(m.ChannelID)
		err := chat.HandleNewMessage(m)
		if err != nil {
			log.Error("handle new message error", zap.Error(err))
		}
	} else {
		log.Info("message is not worthy of reply", zap.String("content", m.Content))
	}
}
