package trivia

import (
	"embed"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io/fs"
	"src-go/util"
)

//go:embed static
var Static embed.FS

func GetQuestions() (result []Question) {
	matches, err := fs.Glob(Static, "**/*.json")

	if err != nil {
		log.Error("failed to get questions", zap.Error(err))

		return result
	}

	if matches == nil || len(matches) == 0 {
		return result
	}

	file := util.RandomElement(matches)
	questions, err := fs.ReadFile(Static, file)
	if err != nil {
		log.Error("failed to read questions", zap.Error(err))
		return result
	}

	_ = json.Unmarshal(questions, &result)
	return result
}

func GetVoice(sentence string, dir string) ([]byte, error) {
	return GetVoiceByHash(util.Hash(sentence), dir)
}

func GetVoiceByHash(hash string, dir string) ([]byte, error) {
	files, err := fs.Glob(Static, fmt.Sprintf("static/%s/%s/*.dca", dir, hash))
	if err != nil {
		return nil, err
	}

	pickedFile := util.RandomElement(files)
	contents, err := fs.ReadFile(Static, pickedFile)
	if err != nil {
		return nil, err
	}

	return contents, nil
}
