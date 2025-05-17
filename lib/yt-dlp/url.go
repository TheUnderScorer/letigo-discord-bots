package ytdlp

import (
	"errors"
	"fmt"
	"net/url"
)

const videoIdQp = "v"

// SanitizeVideoUrl extracts and validates the video ID from a YouTube URL and returns a sanitized URL or an error if invalid.
func SanitizeVideoUrl(videoUrl string) (string, error) {
	u, err := url.Parse(videoUrl)
	if err != nil {
		return "", err
	}

	videoId := u.Query().Get(videoIdQp)
	if videoId == "" {
		return "", errors.New("video id not found")
	}

	return fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoId), nil
}
