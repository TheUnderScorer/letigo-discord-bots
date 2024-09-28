package aws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"src-go/env"
)

const S3ContextKey = "S3"

type S3 struct {
	S3Client *s3.Client
}

func NewS3(client *s3.Client) *S3 {
	return &S3{
		S3Client: client,
	}
}

func (s *S3) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	result, err := s.S3Client.GetObject(ctx, &s3.GetObjectInput{
		Key:    &key,
		Bucket: &env.Env.S3Bucket,
	})
	if err != nil {
		return nil, err
	}

	return result.Body, nil
}
