package ytdlp

import (
	"context"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"os"
	"path"
)

// DownloadOpus downloads audio from the provided URL in Opus format, saves it temporarily, and returns its contents as bytes.
// Returns an error if the download, file reading, or deletion process fails.
func DownloadOpus(ctx context.Context, url string) ([]byte, error) {
	tempDir := os.TempDir()

	fileNameUUID, err := uuid.NewUUID()
	if err != nil {
		log.Error("failed to generate file name", zap.Error(err))
		return nil, err
	}
	fileNameUUIdStr := fileNameUUID.String()
	log.Debug("generated file name", zap.String("name", fileNameUUIdStr))

	fileName := fileNameUUIdStr + ".opus"
	filePath := path.Join(tempDir, fileName)

	cmd := getCommand(ctx, url,
		"--audio-format", "wav",
		"--format", "bestaudio[ext=m4a]/bestaudio/best",
		"-o", fileName,
	)
	cmd.Dir = tempDir
	log.Debug("downloading audio", zap.String("command", cmd.String()))

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error("failed to download audio", zap.Error(err), zap.String("output", string(output)))
		return nil, err
	}
	log.Debug("audio downloaded", zap.String("output", string(output)))

	contents, err := os.ReadFile(filePath)
	if err != nil {
		log.Error("failed to read audio file", zap.Error(err))
		return nil, err
	}

	err = os.Remove(filePath)
	if err != nil {
		log.Error("failed to delete audio file", zap.Error(err))
		return nil, err
	}

	return contents, nil
}
