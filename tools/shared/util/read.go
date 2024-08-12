package util

import (
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/log"
	"os"
	"path/filepath"
)

func ReadStepResult[T any](step string) (out []*T) {
	files, err := filepath.Glob(fmt.Sprintf("%s/result/result-*.json", step))
	if err != nil {
		log.Fatal("failed to read files", err)
	}

	for _, fileName := range files {
		contents, err := os.ReadFile(fileName)
		if err != nil {
			log.Error("failed to read file", err)

			continue
		}

		var v []*T
		err = json.Unmarshal(contents, &v)
		if err != nil {
			log.Error("failed to unmarshal file", err)

			continue
		}

		out = append(out, v...)
	}

	return out
}
