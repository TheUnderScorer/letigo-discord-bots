package chat

import (
	"app/discord"
	"app/llm"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"strings"
)

func HandleMessageCreate(session *discordgo.Session, manager *Manager, llmApi *llm.API, newMessage *discordgo.MessageCreate) {
	log := chatLog.With(zap.String("messageID", newMessage.ID))
	log.Debug("handling message create")

	channel, err := session.Channel(newMessage.ChannelID)
	if err != nil {
		log.Error("failed to get message channel", zap.Error(err))
		return
	}

	sessionUserID := session.State.User.ID

	if newMessage.Author.ID == sessionUserID {
		log.Debug("skip this message, because it was sent by us")

		return
	}

	if manager.HasChat(newMessage.ChannelID) {
		log.Debug("already have chat", zap.String("channelID", newMessage.ChannelID), zap.String("messageID", newMessage.ID))

		doHandleNewMessage(log, session, newMessage, manager)

		return
	}

	// Check if this channel is a thread, and if it was started by us before
	isOurThread := channel.IsThread() && channel.OwnerID == sessionUserID
	if isOurThread {
		log.Debug("channel is a thread started by us before")
	}

	// Check if message explicitly mentions us
	isMention := (newMessage.Mentions != nil && len(newMessage.Mentions) > 0 && discord.HasUser(newMessage.Mentions, sessionUserID)) || strings.Contains(newMessage.Content, sessionUserID)
	if isMention {
		log.Debug("message mentions us explicitly")
	}

	if isOurThread || isMention ||
		// Otherwise, check if message is worthy of reply
		IsWorthyOfReply(llmApi, newMessage.Content) {
		log.Info("message is worthy of reply", zap.String("content", newMessage.Content))
		doHandleNewMessage(log, session, newMessage, manager)
	} else {
		log.Info("message is not worthy of reply", zap.String("content", newMessage.Content))
	}
}

func doHandleNewMessage(log *zap.Logger, session *discordgo.Session, newMessage *discordgo.MessageCreate, manager *Manager) {
	chat := manager.GetOrCreateChat(newMessage.ChannelID)
	err := chat.HandleNewMessage(newMessage)
	if err != nil {
		log.Error("handle new message error", zap.Error(err))
		discord.ReportErrorChannel(session, newMessage.ChannelID, err)
	}
}
