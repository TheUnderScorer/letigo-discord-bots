package main

import (
	"app/aws"
	"app/domain/trivia"
	"app/env"
	util2 "app/util"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/charmbracelet/log"
	"github.com/joho/godotenv"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"tools/shared/util"
)

var ignoredFiles = []string{
	".DS_Store",
	".gitkeep",
}

type fileWalk chan string

func (f fileWalk) Walk(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if !info.IsDir() && !util2.Includes(ignoredFiles, info.Name()) {
		f <- path
	}
	return nil
}

func main() {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Hour*3))
	defer cancel()

	err := godotenv.Load("../.env")
	if err != nil {
		log.Warn("error loading .env file", err)
	}

	env.Init()

	cfg, err := aws.NewConfig(ctx)
	if err != nil {
		log.Fatal(err)
	}

	s3Client := s3.NewFromConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	uploader := manager.NewUploader(s3Client)

	copyQuestions(ctx, uploader)
	copyVoices(ctx, uploader)
}

func copyQuestions(ctx context.Context, uploader *manager.Uploader) {
	questions := util.ReadStepResult[trivia.Question]("3-questions-enhancer")
	contents, err := json.Marshal(questions)
	if err != nil {
		log.Fatal(err)
	}

	hash := util2.Hash(string(contents))

	key := fmt.Sprintf("trivia/questions-%s.json", hash)

	doUpload(ctx, key, bytes.NewBuffer(contents), uploader)
}

func copyVoices(ctx context.Context, uploader *manager.Uploader) {
	walker := make(fileWalk)
	go func() {
		// Gather the files to upload by walking the path recursively
		if err := filepath.Walk("4-questions-voice-generator/result", walker.Walk); err != nil {
			log.Fatal("Walk failed:", err)
		}
		close(walker)
	}()

	for filePath := range walker {
		doUploadPath(ctx, filePath, uploader)
	}
}

func doUploadPath(ctx context.Context, filePath string, uploader *manager.Uploader) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("failed to open file '%s': %v", filePath, err)
	}
	defer file.Close()

	key := filepath.Join("trivia", strings.Replace(filePath, "4-questions-voice-generator/result/", "", 1))
	doUpload(ctx, key, file, uploader)
}

func doUpload(ctx context.Context, key string, file io.Reader, uploader *manager.Uploader) {
	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: &env.Env.S3Bucket,
		Key:    &key,
		Body:   file,
	})
	if err != nil {
		log.Fatalf("failed to upload file '%s': %v", key, err)
	}
}
