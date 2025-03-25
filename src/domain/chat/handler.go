package chat

import (
	gptserverclient "app/llm"
	"context"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

func HandleMessageCreate(ctx context.Context, m *discordgo.MessageCreate) {
	log.Debug("handling message create")
	manager := ctx.Value(ManagerContextKey).(*Manager)

	if manager.HasChat(m.ChannelID) {
		// Get chat, handle message and send reply
		return
	}

	client := ctx.Value(gptserverclient.GtpServerClientContextKey).(*gptserverclient.API)

	if IsWorthyOfReply(client, m.Content) {
		log.Info("message is worthy of reply", zap.String("content", m.Content))
	} else {
		log.Info("message is not worthy of reply", zap.String("content", m.Content))
	}
}
