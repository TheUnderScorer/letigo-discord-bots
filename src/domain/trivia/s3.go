package trivia

import (
	"app/aws"
	"app/util"
	"context"
	"encoding/json"
	"fmt"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
	"io"
	"path/filepath"
	"time"
)

func GetQuestions(s3 *aws.S3) (result []Question) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	bucket := s3.Bucket()
	prefix := filepath.Join("trivia", "questions")
	response, err := s3.Client.ListObjectsV2(ctx, &awss3.ListObjectsV2Input{
		Bucket: &bucket,
		Prefix: &prefix,
	})
	if err != nil {
		log.Error("failed to get questions", zap.Error(err))

		return result
	}

	// Pick two random question.json files from the list
	objects := util.RandomElements(response.Contents, 2)

	for _, object := range objects {
		contents, err := s3.Get(ctx, *object.Key)
		if err != nil {
			log.Error("failed to read contents", zap.String("key", *object.Key), zap.Error(err))
			continue
		}

		var questions []Question
		err = json.NewDecoder(contents).Decode(&questions)
		if err != nil {
			log.Error("failed to decode questions", zap.String("key", *object.Key), zap.Error(err))
		}

		result = append(result, questions...)
	}

	return util.Shuffle(result)
}

func GetVoice(s3 *aws.S3, sentence string) (io.ReadCloser, error) {
	return GetVoiceByHash(s3, util.Hash(sentence))
}

func GetVoiceByHash(s3 *aws.S3, hash string) (io.ReadCloser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	i := util.RandomInt(0, 3)
	key := fmt.Sprintf("trivia/%s/%d.dca", hash, i)
	contents, err := s3.Get(ctx, key)

	if err != nil {
		return nil, err
	}

	return contents, nil
}
