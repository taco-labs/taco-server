package service

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/coolsms/coolsms-go"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
)

var (
	codes = []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}
)

type SmsVerificationSenderService interface {
	GenerateCode(int) string
	SendSmsVerification(context.Context, entity.SmsVerification) error
}

type mockSmsSenderService struct{}

func (m mockSmsSenderService) GenerateCode(length int) string {
	b := make([]string, length)
	for i := range b {
		b[i] = "1"
	}
	return strings.Join(b, "")
}

func (m mockSmsSenderService) SendSmsVerification(ctx context.Context, smsVerification entity.SmsVerification) error {
	return nil
}

func NewMockSmsSenderService() mockSmsSenderService {
	return mockSmsSenderService{}
}

type coolSmsSenderService struct {
	phoneFrom string
	client    *coolsms.Client
}

func (s coolSmsSenderService) GenerateCode(length int) string {
	b := make([]string, length)
	for i := range b {
		b[i] = codes[rand.Intn(len(codes))]
	}
	return strings.Join(b, "")

}

func (s coolSmsSenderService) SendSmsVerification(ctx context.Context, smsVerification entity.SmsVerification) error {
	msg := make(map[string]interface{})

	msg["to"] = smsVerification.Phone
	msg["from"] = s.phoneFrom
	msg["type"] = "SMS"
	msg["text"] = fmt.Sprintf("[타코택시] 인증 코드 [%s]를 입력해주세요.", smsVerification.VerificationCode)

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
