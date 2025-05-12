package ytdlp

import (
	"context"
	"go.uber.org/zap"
	"strings"
)

// GetTitle retrieves the title of a video from a given URL using the yt-dlp command-line tool.
func GetTitle(ctx context.Context, url string) (string, error) {
	cmd := getCommand(ctx, url, "--get-title")

	result, err := cmd.CombinedOutput()
	if err != nil {
		log.Error("failed to get title", zap.Error(err), zap.ByteString("output", result))
		return "", err
	}

	title := strings.TrimSpace(string(result))
	log.Debug("title retrieved", zap.String("title", title))
	return title, nil
}
