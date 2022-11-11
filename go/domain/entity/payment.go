package entity

import (
	"time"

	"github.com/uptrace/bun"
)

type UserPayment struct {
	bun.BaseModel `bun:"table:user_payment"`

	Id                  string    `bun:"id,pk"`
	UserId              string    `bun:"user_id"`
	Name                string    `bun:"name"`
	CardCompany         string    `bun:"card_company"`
	RedactedCardNumber  string    `bun:"redacted_card_number"`
	CardExpirationYear  string    `bun:"card_expiration_year"`
	CardExpirationMonth string    `bun:"card_expiration_month"`
	BillingKey          string    `bun:"billing_key"`
	DefaultPayment      bool      `bun:"default_payment"`
	CreateTime          time.Time `bun:"create_time"`
}

type UserDefaultPayment struct {
	bun.BaseModel `bun:"table:user_default_payment"`

	UserId    string `bun:"user_id,pk"`
	PaymentId string `bun:"payment_id"`
}
