package service

import (
	"context"
	"net/url"
)

type ImageDownloadUrlService interface {
	GetDownloadUrl(context.Context, string) (string, error)
}

type s3PublicAccessUrlService struct {
	domainName string
	basePath   string
}

func (s s3PublicAccessUrlService) GetDownloadUrl(ctx context.Context, path string) (string, error) {
	return url.JoinPath(s.domainName, s.basePath, path)
}

func NewS3PublicAccessUrlService(domainName string, basePath string) *s3PublicAccessUrlService {
	return &s3PublicAccessUrlService{
		domainName: domainName,
		basePath:   basePath,
	}
}
