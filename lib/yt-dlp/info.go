package ytdlp

import (
	"go.uber.org/zap"
	"strings"
)

// GetTitle retrieves the title of a video from a given URL using the yt-dlp command-line tool.
func GetTitle(url string) (string, error) {
	cmd := getCommand(url, "--get-title")

	result, err := cmd.CombinedOutput()
	if err != nil {
		log.Error("failed to get title", zap.Error(err), zap.ByteString("output", result))
		return "", err
	}

	return strings.TrimSpace(string(result)), nil
}
