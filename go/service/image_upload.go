package service

import (
	"context"
	"os"
)

type FileUploadService interface {
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
