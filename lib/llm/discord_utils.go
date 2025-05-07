package llm

import (
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"io"
	"lib/logging"
	"net/http"
)

// HandleDiscordMessageAttachments processes attachments in a Discord message and returns them as a slice of File objects.
func HandleDiscordMessageAttachments(message *discordgo.Message) []File {
	log := logging.Get().Named("llm").Named("DiscordUtils")
	files := make([]File, 0)
	if len(message.Attachments) > 0 {
		for _, attachment := range message.Attachments {
			response, err := http.DefaultClient.Get(attachment.URL)
			if err != nil {
				log.Error("failed to download attachment", zap.Error(err))
				continue
			}
			data, err := io.ReadAll(response.Body)
			if err != nil {
				log.Error("failed to read attachment", zap.Error(err))
				continue
			}

			files = append(files, File{
				Data:        data,
				Name:        attachment.Filename,
				ContentType: attachment.ContentType,
			})
		}
	}
	return files
}
