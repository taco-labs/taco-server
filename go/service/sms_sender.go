package service

import (
	"context"
	"fmt"

	"math/rand"
	"strings"

	"github.com/coolsms/coolsms-go"
	"github.com/taco-labs/taco/go/domain/value"
)

var (
	codes = []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}
)

type SmsVerificationSenderService interface {
	SendSmsVerification(context.Context, string) (string, error)
}

type mockSmsSenderService struct{}

func (m mockSmsSenderService) SendSmsVerification(ctx context.Context, phone string) (string, error) {
	return "111111", nil
}

func NewMockSmsSenderService() mockSmsSenderService {
	return mockSmsSenderService{}
}

type coolSmsSenderService struct {
	phoneFrom string
	client    *coolsms.Client
}

func (s coolSmsSenderService) SendSmsVerification(ctx context.Context, phone string) (string, error) {
	verificationCode := generateRandomCode(6)

	msg := make(map[string]interface{})

	msg["to"] = phone
	msg["from"] = s.phoneFrom
	msg["type"] = "SMS"
	msg["text"] = fmt.Sprintf("[타코] 인증 코드 [%s]를 입력해주세요.", verificationCode)

	params := make(map[string]interface{})
	params["message"] = msg

	_, err := s.client.Messages.SendSimpleMessage(params)
	if err != nil {
		return "", fmt.Errorf("%w: %v", value.ErrExternal, err)
	}
	return verificationCode, nil
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

func generateRandomCode(length int) string {
	b := make([]string, length)
	for i := range b {
		b[i] = codes[rand.Intn(len(codes))]
	}
	return strings.Join(b, "")
}
