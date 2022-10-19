package entity

import (
	"time"

	"github.com/taco-labs/taco/go/domain/value/enum"
	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:\"user\""`

	Id            string      `bun:"id,pk"`
	FirstName     string      `bun:"first_name"`
	LastName      string      `bun:"last_name"`
	BirthDay      string      `bun:"birthday"`
	Phone         string      `bun:"phone"`
	Gender        string      `bun:"gender"`
	AppOs         enum.OsType `bun:"app_os"`
	AppVersion    string      `bun:"app_version"`
	UserUniqueKey string      `bun:"user_unique_key"`
	CreateTime    time.Time   `bun:"create_time"`
	UpdateTime    time.Time   `bun:"update_time"`
	DeleteTime    time.Time   `bun:"delete_time"`
}

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
