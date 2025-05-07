package aws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
)

type S3 struct {
	Client *s3.Client
	Bucket string
}

func NewS3(client *s3.Client, bucket string) *S3 {
	return &S3{
		Client: client,
		Bucket: bucket,
	}
}

func (s *S3) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	bucket := s.Bucket
	result, err := s.Client.GetObject(ctx, &s3.GetObjectInput{
		Key:    &key,
		Bucket: &bucket,
	})
	if err != nil {
		return nil, err
	}

	return result.Body, nil
}
