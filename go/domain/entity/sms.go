package entity

import (
	"time"

	"github.com/uptrace/bun"
)

var (
	MockAccountPhone = "01000000000"
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
	return s.Phone == MockAccountPhone
}

func NewSmsVerification(id string, verificationCode string, currentTime time.Time, phone string) SmsVerification {
	return SmsVerification{
		Id:               id,
		VerificationCode: verificationCode,
		Verified:         false,
		ExpireTime:       currentTime.Add(time.Minute * 3),
		Phone:            phone,
	}
}

func NewMockSmsVerification(id string, currentTime time.Time) SmsVerification {
	return SmsVerification{
		Id:               id,
		VerificationCode: "000000",
		Verified:         false,
		ExpireTime:       currentTime.Add(time.Minute * 3),
		Phone:            MockAccountPhone,
	}
}
