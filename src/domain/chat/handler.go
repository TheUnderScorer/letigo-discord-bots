package chat

import (
	"app/llm"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

func HandleMessageCreate(manager *Manager, llmApi *llm.API, m *discordgo.MessageCreate) {
	log.Debug("handling message create")

	if manager.HasChat(m.ChannelID) {
		// Get chat, handle message and send reply
		return
	}

	if IsWorthyOfReply(llmApi, m.Content) {
		log.Info("message is worthy of reply", zap.String("content", m.Content))
	} else {
		log.Info("message is not worthy of reply", zap.String("content", m.Content))
	}
}
