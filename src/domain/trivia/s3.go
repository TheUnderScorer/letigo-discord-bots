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
)

func GetQuestions(ctx context.Context) (result []Question) {
	s3 := ctx.Value(aws.S3ContextKey).(*aws.S3)
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

func GetVoice(ctx context.Context, sentence string) (io.ReadCloser, error) {
	return GetVoiceByHash(ctx, util.Hash(sentence))
}

func GetVoiceByHash(ctx context.Context, hash string) (io.ReadCloser, error) {
	s3 := ctx.Value(aws.S3ContextKey).(*aws.S3)

	i := util.RandomInt(0, 3)
	key := fmt.Sprintf("trivia/%s/%d.dca", hash, i)
	contents, err := s3.Get(ctx, key)

	if err != nil {
		return nil, err
	}

	return contents, nil
}
