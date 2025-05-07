package youtube

import (
	"errors"
	yt "github.com/kkdai/youtube/v2"
	"go.uber.org/zap"
	"lib/logging"
	"strings"
)

var logger = logging.Get().Named("youtube")

var ytClient = yt.Client{}

func findBestFormat(formats yt.FormatList) *yt.Format {
	for _, format := range formats {
		if strings.Contains(format.MimeType, "opus") && strings.Contains(format.MimeType, "audio") {
			return &format
		}
	}

	fallback := formats[0]
	logger.Warn("opus format not found, using first one as fallback", zap.Any("format", fallback))

	return &fallback
}

func GetAudioURL(videoID string) (string, *yt.Format, error) {
	video, err := ytClient.GetVideo(videoID)
	if err != nil {
		return "", nil, err
	}

	formats := video.Formats.WithAudioChannels()
	format := findBestFormat(formats)
	if format == nil {
		return "", nil, errors.New("no opus audio format found")
	}
	logger.Info("got format", zap.Any("format", format), zap.String("videoID", videoID))

	stream, err := ytClient.GetStreamURL(video, format)
	if err != nil {
		return "", nil, err
	}
	return stream, format, nil
}
