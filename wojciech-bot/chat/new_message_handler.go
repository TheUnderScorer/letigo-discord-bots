package chat

import (
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"lib/discord"
	"strings"
)

func HandleMessageCreate(bot *discord.Bot, manager *Manager, newMessage *discordgo.MessageCreate) {
	log := chatLog.With(zap.String("messageID", newMessage.ID))
	log.Info("handling message create")

	channel, err := bot.Channel(newMessage.ChannelID)
	if err != nil {
		log.Error("failed to get message channel", zap.Error(err))
		return
	}

	sessionUserID := bot.State.User.ID

	if newMessage.Author.ID == sessionUserID {
		log.Debug("skip this message, because it was sent by us")

		return
	}

	if manager.HasChat(newMessage.ChannelID) {
		log.Debug("already have chat", zap.String("channelID", newMessage.ChannelID), zap.String("messageID", newMessage.ID))

		doHandleNewMessage(log, bot, newMessage, manager)

		return
	}

	isDM := channel.Type == discordgo.ChannelTypeDM
	if isDM {
		log.Debug("channel is DM")
	}

	// Check if this channel is a thread and if it was started by us before
	isOurThread := channel.IsThread() && channel.OwnerID == sessionUserID
	if isOurThread {
		log.Debug("channel is a thread started by us before")
	}

	// Check if a message explicitly mentions us
	isMention := (len(newMessage.Mentions) > 0 && discord.HasUser(newMessage.Mentions, sessionUserID)) || strings.Contains(newMessage.Content, sessionUserID)
	if isMention {
		log.Debug("message mentions us explicitly")
	}

	if isOurThread || isMention || isDM {
		log.Info("message is worthy of reply", zap.String("content", newMessage.Content))
		doHandleNewMessage(log, bot, newMessage, manager)
	} else {
		log.Info("message is not worthy of reply", zap.String("content", newMessage.Content))
	}
}

func doHandleNewMessage(log *zap.Logger, bot *discord.Bot, newMessage *discordgo.MessageCreate, manager *Manager) {
	chat := manager.GetOrCreateChat(newMessage.ChannelID)
	err := chat.HandleNewMessage(newMessage.Message)
	if err != nil {
		log.Error("handle new message error", zap.Error(err))
		bot.ReportErrorChannel(newMessage.ChannelID, err)
	}
}
