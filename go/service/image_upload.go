package service

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/taco-labs/taco/go/domain/value"
)

type ImageUploaderService interface {
	Upload(context.Context, os.File) (string, error)
	Delete(context.Context, string) error
}

type mockFileUploadService struct{}

func (m mockFileUploadService) Upload(_ context.Context, _ os.File) (string, error) {
	return "/testurl", nil
}

func (m mockFileUploadService) Delete(_ context.Context, _ string) error {
	return nil
}

func NewMockFileUploadService() mockFileUploadService {
	return mockFileUploadService{}
}

type ImageUrlService interface {
	GetDownloadUrl(context.Context, string) (string, error)
	GetUploadUrl(context.Context, string) (string, error)
}

type s3PresignedUrlService struct {
	client   *s3.PresignClient
	timeout  time.Duration
	bucket   string
	basePath string
}

func (s s3PresignedUrlService) GetDownloadUrl(ctx context.Context, path string) (string, error) {
	presignParams := s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    aws.String(getS3Key(s.basePath, path)),
	}

	presignDuration := func(po *s3.PresignOptions) {
		po.Expires = s.timeout
	}

	presignResult, err := s.client.PresignGetObject(ctx, &presignParams, presignDuration)

	if err != nil {
		return "", fmt.Errorf("%w: error while presign get object request: %v", value.ErrExternal, err)
	}

	return presignResult.URL, nil
}

func (s s3PresignedUrlService) GetUploadUrl(ctx context.Context, path string) (string, error) {
	presignParams := s3.PutObjectInput{
		Bucket: &s.bucket,
		Key:    aws.String(getS3Key(s.basePath, path)),
	}

	presignDuration := func(po *s3.PresignOptions) {
		po.Expires = s.timeout
	}

	presignResult, err := s.client.PresignPutObject(ctx, &presignParams, presignDuration)

	if err != nil {
		return "", fmt.Errorf("%w: error while presign put object request: %v", value.ErrExternal, err)
	}

	return presignResult.URL, nil
}

func NewS3ImagePresignedUrlService(client *s3.PresignClient, timeout time.Duration, bucket string, basePath string) s3PresignedUrlService {
	return s3PresignedUrlService{
		client:   client,
		timeout:  timeout,
		bucket:   bucket,
		basePath: basePath,
	}
}

func getS3Key(basePath string, path string) string {
	return fmt.Sprintf("%s/%s", basePath, path)
}
