package entity

import (
	"time"

	"github.com/taco-labs/taco/go/domain/value"
	"github.com/uptrace/bun"
)

type SmsVerification struct {
	bun.BaseModel `bun:"table:sms_verification"`

	Id               string    `bun:"id,pk"`
	VerificationCode string    `bun:"verification_code"`
	Verified         bool      `bun:"verified"`
	Phone            string    `bun:"phone"`
	ExpireTime       time.Time `bun:"expire_time"`
}

func (s SmsVerification) MockAccountPhone() bool {
	return value.IsMockPhoneNumber(s.Phone)
}

func NewSmsVerification(id, code string, currentTime time.Time, phone string) SmsVerification {
	return SmsVerification{
		Id:               id,
		VerificationCode: code,
		Verified:         false,
		ExpireTime:       currentTime.Add(time.Minute * 3),
		Phone:            phone,
	}
}

func NewMockSmsVerification(id string, currentTime time.Time, phone string) SmsVerification {
	return SmsVerification{
		Id:               id,
		VerificationCode: "000000",
		Verified:         false,
		ExpireTime:       currentTime.Add(time.Minute * 3),
		Phone:            phone,
	}
}
