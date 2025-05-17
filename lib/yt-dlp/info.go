package ytdlp

import (
	"context"
	"go.uber.org/zap"
	"lib/errors"
	"strconv"
	"strings"
	"time"
)

type VideoMetadata struct {
	Title        string
	Duration     time.Duration
	ThumbnailUrl string
}

const empty = "NA"

// GetMetadata retrieves the metadata of a video from a given URL using the yt-dlp command-line tool.
func GetMetadata(ctx context.Context, url string) (*VideoMetadata, error) {
	parsedUrl, err := SanitizeVideoUrl(url)
	if err != nil {
		return nil, err
	}

	cmd := getCommand(ctx, parsedUrl, "--print", "duration,title,thumbnail")

	result, err := cmd.CombinedOutput()
	if err != nil {
		log.Error("failed to get output", zap.Error(err), zap.ByteString("output", result))
		return nil, err
	}

	output := strings.TrimSpace(string(result))
	log.Debug("output retrieved", zap.String("output", output))

	outputParts := strings.Split(output, "\n")
	duration, err := strconv.Atoi(outputParts[0])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse duration")
	}
	log.Debug("duration parsed", zap.Int("duration", duration))
	title := outputParts[1]
	thumbnailUrl := outputParts[2]
	if thumbnailUrl == empty {
		thumbnailUrl = ""
	}

	timeDuration := time.Duration(duration) * time.Second
	metadata := &VideoMetadata{
		Title:        title,
		Duration:     timeDuration,
		ThumbnailUrl: thumbnailUrl,
	}
	log.Debug("parsed duration", zap.Duration("duration", timeDuration))

	return metadata, nil
}
