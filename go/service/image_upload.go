package service

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/eko/gocache/v3/cache"
	"github.com/taco-labs/taco/go/domain/value"
)

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

func NewS3ImagePresignedUrlService(client *s3.PresignClient, timeout time.Duration, bucket string, basePath string) *s3PresignedUrlService {
	return &s3PresignedUrlService{
		client:   client,
		timeout:  timeout,
		bucket:   bucket,
		basePath: basePath,
	}
}

func getS3Key(basePath string, path string) string {
	return fmt.Sprintf("%s/%s", basePath, path)
}

type cachedUrlService struct {
	svc              ImageUrlService
	downloadUrlCache cache.CacheInterface[string]
	uploadUrlCache   cache.CacheInterface[string]
}

func (c cachedUrlService) GetDownloadUrl(ctx context.Context, key string) (string, error) {
	return c.downloadUrlCache.Get(ctx, key)
}

func (c cachedUrlService) GetUploadUrl(ctx context.Context, key string) (string, error) {
	return c.uploadUrlCache.Get(ctx, key)
}

func NewCachedUrlService(
	downloadCacheInterface cache.CacheInterface[string],
	uploadCacheInterface cache.CacheInterface[string],
	svc ImageUrlService) *cachedUrlService {

	loadDownloadUrlFn := func(ctx context.Context, key any) (string, error) {
		keyStr, ok := key.(string)
		if !ok {
			return "", fmt.Errorf("%w: Can not convert key (%v) to string", value.ErrInternal, key)
		}
		return svc.GetDownloadUrl(ctx, keyStr)
	}

	loadUploadUrlFn := func(ctx context.Context, key any) (string, error) {
		keyStr, ok := key.(string)
		if !ok {
			return "", fmt.Errorf("%w: Can not convert key (%v) to string", value.ErrInternal, key)
		}
		return svc.GetUploadUrl(ctx, keyStr)
	}

	downloadCache := cache.NewLoadable(loadDownloadUrlFn, downloadCacheInterface)
	uploadCache := cache.NewLoadable(loadUploadUrlFn, uploadCacheInterface)

	return &cachedUrlService{
		svc:              svc,
		downloadUrlCache: downloadCache,
		uploadUrlCache:   uploadCache,
	}
}
