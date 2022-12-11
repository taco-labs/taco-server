package entity

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/uptrace/bun"
)

var (
	codes = []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}
)

type SmsVerification struct {
	bun.BaseModel `bun:"table:sms_verification"`

	Id               string    `bun:"id,pk"`
	VerificationCode string    `bun:"verification_code"`
	Verified         bool      `bun:"verified"`
	Phone            string    `bun:"phone"`
	ExpireTime       time.Time `bun:"expire_time"`
}

func (s SmsVerification) VerficationMessage() string {
	return fmt.Sprintf("[타코] 인증 코드 [%s]를 입력해주세요.", s.VerificationCode)
}

func NewSmsVerification(id string, currentTime time.Time, phone string) SmsVerification {
	return SmsVerification{
		Id:               id,
		VerificationCode: generateRandomCode(6),
		Verified:         false,
		ExpireTime:       currentTime.Add(time.Minute * 3),
		Phone:            phone,
	}
}

func generateRandomCode(length int) string {
	b := make([]string, length)
	for i := range b {
		b[i] = codes[rand.Intn(len(codes))]
	}
	return strings.Join(b, "")
}
