package chat

import (
	"context"
	"errors"
	"github.com/bwmarrin/discordgo"
	"github.com/openai/openai-go"
	"go.uber.org/zap"
	"lib/discord"
	apperrors "lib/errors"
	"lib/logging"
)

type ForgetComponentHandler struct {
	openaiClient *openai.Client
}

func NewForgetComponentHandler(openaiClient *openai.Client) ForgetComponentHandler {
	return ForgetComponentHandler{
		openaiClient: openaiClient,
	}
}

var logger = logging.Get()

func (f ForgetComponentHandler) Handle(ctx context.Context, interaction *discordgo.InteractionCreate, bot *discord.Bot) error {
	log := logger.With(zap.String("interactionID", interaction.ID))

	log.Info("handling forget operation")

	var vectorFileID string
	var vectorStoreID string
	var fileID string

	for _, embed := range interaction.Message.Embeds {
		if len(embed.Fields) > 0 {
			for _, field := range embed.Fields {
				switch field.Name {
				case MemoryEmbedFieldVectorStoreID.String():
					vectorStoreID = field.Value

				case MemoryEmbedFieldVectorFileID.String():
					vectorFileID = field.Value

				case MemoryEmbedFileID.String():
					fileID = field.Value
				}
			}
		}
	}

	if vectorFileID == "" {
		return errors.New("no vector file id found")
	}

	if vectorStoreID == "" {
		return errors.New("no vector store id found")
	}

	if fileID == "" {
		return errors.New("no file id found")
	}

	_, err := f.openaiClient.VectorStores.Files.Get(ctx, vectorStoreID, vectorFileID)
	if err != nil {
		return apperrors.Wrap(err, "failed to get vector file")
	}

	_, err = f.openaiClient.VectorStores.Files.Delete(ctx, vectorStoreID, vectorFileID)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete vector file")
	}

	_, err = f.openaiClient.Files.Delete(ctx, fileID)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete file")
	}

	components := ForgetMessageComponent(true)
	_, err = bot.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:         interaction.Message.ID,
		Channel:    interaction.ChannelID,
		Embeds:     &interaction.Message.Embeds,
		Components: &components,
	})
	if err != nil {
		return apperrors.Wrap(err, "failed to edit message")
	}

	log.Info("memory forgotten", zap.String("vectorFileID", vectorFileID), zap.String("vectorStoreID", vectorStoreID), zap.String("fileID", fileID))

	return nil
}

func (f ForgetComponentHandler) ShouldHandle(interaction *discordgo.InteractionCreate) bool {
	return interaction.MessageComponentData().CustomID == ForgetButtonID && interaction.Message != nil
}
