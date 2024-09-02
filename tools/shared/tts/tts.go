package tts

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"src-go/domain/tts"
)

type Manifest map[string]ManifestEntry

type ManifestEntry struct {
	// FileName contains file path relative to outDir
	FileName string `json:"file_name"`
}

// GenerateSentences generates TTS audio for given sentences using typer script
func GenerateSentences(ctx context.Context, speaker tts.Speaker, outDir string) error {
	dir, _ := os.Getwd()
	cmd := exec.CommandContext(ctx, "poetry", fmt.Sprintf("run w-tts --speaker %s --out_dir %s", speaker, outDir))
	cmd.Path = path.Join(dir, "../", "w-tts")

	err := cmd.Run()
	if err != nil {
		return err
	}

	var manifest map[string]Manifest

	file, err := os.Open(fmt.Sprintf("%s/.manifest.json", outDir))
	if err != nil {
		return err
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&manifest)
	return err
}
