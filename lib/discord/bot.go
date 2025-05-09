package discord

import (
	goerrors "errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"lib/errors"
	"lib/logging"
	"lib/util/arrayutil"
)

type BotMessages struct {
	UnknownError string
}

// Bot represents a wrapper around a Discord session to simplify interaction operations and messaging functionality.
type Bot struct {
	*discordgo.Session
	Name     string
	Messages BotMessages
}

// NewBot creates and returns a new Bot instance using the provided Discord bot token. Logs and exits if the token is empty.
func NewBot(token string, name string, messages BotMessages) *Bot {
	if token == "" {
		log.Fatal("token is empty")
	}

	session := NewClient(token)

	return &Bot{
		Session:  session,
		Name:     name,
		Messages: messages,
	}
}

// InteractionReply represents a response to a Discord interaction event.
// Content is the main text content of the interaction reply.
// Ephemeral determines the visibility of the message (true for private, false for public).
// Embeds is a slice of rich-embed objects included in the interaction reply.
type InteractionReply struct {
	Content   string
	Ephemeral bool
	Embeds    []*discordgo.MessageEmbed
}

// logger is a global instance of *zap.Logger used to log application events and errors. Initialized via logging.Get().
var logger = logging.Get().Named("messages")

// SendMessageAndForget sends a message to the specified channel and logs any errors without returning them to the caller.
func (b *Bot) SendMessageAndForget(channelID string, content string) {
	_, err := b.ChannelMessageSend(channelID, content)
	if err != nil {
		logger.Error("failed to send message", zap.Error(err), zap.String("channelID", channelID))
	}
}

// ReplyToInteractionAndForget sends a response to the given interaction and logs errors if the operation fails.
func (b *Bot) ReplyToInteractionAndForget(i *discordgo.Interaction, reply *InteractionReply) {
	var flags discordgo.MessageFlags
	if reply.Ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}

	err := b.InteractionRespond(i, &discordgo.InteractionResponse{
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

// FollowupInteractionErrorAndForget sends an ephemeral error message as a follow-up to the given interaction and logs failures.
func (b *Bot) FollowupInteractionErrorAndForget(i *discordgo.Interaction, errToSend error) {
	flags := discordgo.MessageFlagsEphemeral

	_, err := b.FollowupMessageCreate(i, false, &discordgo.WebhookParams{
		Flags:   flags,
		Content: errToSend.Error(),
		Embeds:  prepareErrorReportEmbed(errToSend),
	})

	if err != nil {
		logger.Error("failed to respond", zap.Error(err), zap.Any("interaction", i))
	}
}

// RespondToInteractionAndForget sends an interaction response without handling or propagating any errors that may occur.
func (b *Bot) RespondToInteractionAndForget(i *discordgo.Interaction, response *discordgo.InteractionResponse) {
	err := b.InteractionRespond(i, response)
	if err != nil {
		logger.Error("failed to respond", zap.Error(err), zap.Any("interaction", i))
	}
}

// SendMessageComplexAndForget sends a complex message to the specified channel and logs an error if the operation fails.
func (b *Bot) SendMessageComplexAndForget(channelID string, content *discordgo.MessageSend) {
	_, err := b.ChannelMessageSendComplex(channelID, content)
	if err != nil {
		logger.Error("failed to send message", zap.Error(err), zap.String("channelID", channelID))
	}
}

// DeleteFollowupAndForget deletes the follow-up interaction response and logs an error if the operation fails.
func (b *Bot) DeleteFollowupAndForget(i *discordgo.Interaction) {
	err := b.InteractionResponseDelete(i)

	if err != nil {
		logger.Error("failed to delete followup", zap.Error(err), zap.Any("interaction", i))
	}
}

// FollowupInteractionMessageAndForget sends a follow-up message for an interaction and logs errors without returning them.
func (b *Bot) FollowupInteractionMessageAndForget(i *discordgo.Interaction, reply *InteractionReply) {
	var flags discordgo.MessageFlags
	if reply.Ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}

	_, err := b.FollowupMessageCreate(i, false, &discordgo.WebhookParams{
		Flags:   flags,
		Content: reply.Content,
		Embeds:  reply.Embeds,
	})

	if err != nil {
		logger.Error("failed to respond", zap.Error(err), zap.Any("interaction", i))
	}
}

// StartLoadingInteractionAndForget sends a deferred response to the interaction to indicate loading status asynchronously.
// It logs an error if the response fails.
func (b *Bot) StartLoadingInteractionAndForget(i *discordgo.Interaction) {
	err := b.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})

	if err != nil {
		logger.Error("failed to respond", zap.Error(err), zap.Any("interaction", i))
	}
}

func (b *Bot) FollowUpInteractionErrorReply(err error, i *discordgo.Interaction) {
	if err != nil {
		var errPublic *errors.ErrPublic
		if goerrors.As(err, &errPublic) {
			b.FollowupInteractionErrorAndForget(i, errPublic)
		} else {
			b.FollowupInteractionMessageAndForget(i, &InteractionReply{
				Content: b.Messages.UnknownError,
				Embeds:  prepareErrorReportEmbed(err),
			})
		}
	}
}

func (b *Bot) ReportErrorChannel(cid string, err error) {
	embeds := prepareErrorReportEmbed(err)

	b.SendMessageComplexAndForget(cid, &discordgo.MessageSend{
		Content: b.Messages.UnknownError,
		Embeds:  embeds,
	})
}

func (b *Bot) ReportErrorInteraction(interaction *discordgo.InteractionCreate, err error) {
	embeds := prepareErrorReportEmbed(err)

	b.FollowupInteractionMessageAndForget(interaction.Interaction, &InteractionReply{
		Content:   b.Messages.UnknownError,
		Embeds:    embeds,
		Ephemeral: true,
	})
}

func (b *Bot) ListVoiceChannelMembers(gid string, cid string) ([]*discordgo.Member, error) {
	guild, err := b.State.Guild(gid)
	if err != nil {
		return nil, err
	}

	var members []*discordgo.Member
	var ids []string
	for _, member := range guild.VoiceStates {
		if member.ChannelID != cid {
			continue
		}

		ids = append(ids, member.UserID)
	}

	if len(ids) > 0 {
		fetchedMembers, err := b.GuildMembers(guild.ID, "", 1000)
		if err != nil {
			return nil, err
		}

		for _, member := range fetchedMembers {
			if member.User != nil && arrayutil.Includes(ids, member.User.ID) {
				members = append(members, member)
			}
		}
	}

	return members, nil
}

func prepareErrorReportEmbed(err error) []*discordgo.MessageEmbed {
	var publicError *errors.ErrPublic
	fields := make(map[string]string)
	fields["Error"] = err.Error()

	if goerrors.As(err, &publicError) {
		if publicError.Cause != nil {
			publicError.AddContext("cause", publicError.Cause.Error())
		}

		for key, value := range publicError.Context {
			fields[key] = fmt.Sprintf("%v", value)
		}

	}

	var embedFields []*discordgo.MessageEmbedField
	for key, value := range fields {
		embedFields = append(embedFields, &discordgo.MessageEmbedField{
			Name:  key,
			Value: value,
		})
	}

	return []*discordgo.MessageEmbed{
		{
			Title:  "Error details",
			Fields: embedFields,
			Color:  0xff0000,
		},
	}
}
