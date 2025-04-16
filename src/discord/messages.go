package discord

import (
	"app/logging"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type InteractionReply struct {
	Content   string
	Ephemeral bool
	Embeds    []*discordgo.MessageEmbed
}

var logger = logging.Get().Named("messages")

func SendMessageAndForget(s *discordgo.Session, channelID string, content string) {
	_, err := s.ChannelMessageSend(channelID, content)
	if err != nil {
		logger.Error("failed to send message", zap.Error(err), zap.String("channelID", channelID))
	}
}

func SendMessageComplexAndForget(s *discordgo.Session, channelID string, content *discordgo.MessageSend) {
	_, err := s.ChannelMessageSendComplex(channelID, content)
	if err != nil {
		logger.Error("failed to send message", zap.Error(err), zap.String("channelID", channelID))
	}
}

func ReplyToInteractionAndForget(s *discordgo.Session, i *discordgo.Interaction, reply *InteractionReply) {
	var flags discordgo.MessageFlags
	if reply.Ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}

	err := s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: reply.Content,
			Flags:   flags,
		},
	})

	if err != nil {
		logger.Error("failed to respond", zap.Error(err), zap.Any("interaction", i))
	}
}

func DeleteFollowupAndForget(s *discordgo.Session, i *discordgo.Interaction) {
	err := s.InteractionResponseDelete(i)

	if err != nil {
		logger.Error("failed to delete followup", zap.Error(err), zap.Any("interaction", i))
	}
}

func FollowupInteractionErrorAndForget(s *discordgo.Session, i *discordgo.Interaction, errToSend error) {
	flags := discordgo.MessageFlagsEphemeral

	_, err := s.FollowupMessageCreate(i, false, &discordgo.WebhookParams{
		Flags:   flags,
		Content: errToSend.Error(),
	})

	if err != nil {
		logger.Error("failed to respond", zap.Error(err), zap.Any("interaction", i))
	}
}

func RespondToInteractionAndForget(s *discordgo.Session, i *discordgo.Interaction, response *discordgo.InteractionResponse) {
	err := s.InteractionRespond(i, response)
	if err != nil {
		logger.Error("failed to respond", zap.Error(err), zap.Any("interaction", i))
	}
}

func FollowupInteractionMessageAndForget(s *discordgo.Session, i *discordgo.Interaction, reply *InteractionReply) {
	var flags discordgo.MessageFlags
	if reply.Ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}

	_, err := s.FollowupMessageCreate(i, false, &discordgo.WebhookParams{
		Flags:   flags,
		Content: reply.Content,
		Embeds:  reply.Embeds,
	})

	if err != nil {
		logger.Error("failed to respond", zap.Error(err), zap.Any("interaction", i))
	}
}

func StartLoadingInteractionAndForget(s *discordgo.Session, i *discordgo.Interaction) {
	err := s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})

	if err != nil {
		logger.Error("failed to respond", zap.Error(err), zap.Any("interaction", i))
	}
}
