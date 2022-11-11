package service

import (
	"context"
	"fmt"

	"github.com/coolsms/coolsms-go"
	"github.com/taco-labs/taco/go/domain/value"
)

type SmsSenderService interface {
	SendSms(context.Context, string, string) error
}

type mockSmsSenderService struct{}

func (m mockSmsSenderService) SendSms(ctx context.Context, phone string, message string) error {
	return nil
}

func NewMockSmsSenderService() mockSmsSenderService {
	return mockSmsSenderService{}
}

type coolSmsSenderService struct {
	phoneFrom string
	client    *coolsms.Client
}

func (s coolSmsSenderService) SendSms(ctx context.Context, phone string, message string) error {
	msg := make(map[string]interface{})

	msg["to"] = phone
	msg["from"] = s.phoneFrom
	msg["type"] = "SMS"
	msg["text"] = message

	params := make(map[string]interface{})
	params["message"] = msg

	_, err := s.client.Messages.SendSimpleMessage(params)
	if err != nil {
		return fmt.Errorf("%w: %v", value.ErrExternal, err)
	}
	return nil
}

func NewCoolSmsSenderService(endpoint string, phoneFrom string, apiKey string, apiSecret string) *coolSmsSenderService {
	client := coolsms.NewClient()
	client.Messages.Config = map[string]string{
		"APIKey":    apiKey,
		"APISecret": apiSecret,
		"Protocol":  "https",
		"Domain":    endpoint,
	}

	return &coolSmsSenderService{
		client:    client,
		phoneFrom: phoneFrom,
	}
}
