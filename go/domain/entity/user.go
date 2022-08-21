package entity

import (
	"database/sql"
	"time"

	"github.com/ktk1012/taco/go/domain/value"
	"github.com/ktk1012/taco/go/domain/value/enum"
	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:\"user\""`

	Id               string         `bun:"id,pk"`
	FirstName        string         `bun:"first_name"`
	LastName         string         `bun:"last_name"`
	Email            string         `bun:"email"`
	BirthDay         string         `bun:"birthday"`
	Phone            string         `bun:"phone"`
	Gender           string         `bun:"gender"`
	AppOs            enum.OsType    `bun:"app_os"`
	OsVersion        string         `bun:"os_version"`
	AppVersion       string         `bun:"app_version"`
	AppFcmToken      string         `bun:"app_fcm_token"`
	UserUniqueKey    string         `bun:"user_unique_key"`
	DefaultPaymentId sql.NullString `bun:"default_payment_id"`
	CreateTime       time.Time      `bun:"create_time"`
	UpdateTime       time.Time      `bun:"update_time"`
	DeleteTime       time.Time      `bun:"delete_time"`
}

type UserPayment struct {
	bun.BaseModel `bun:"table:user_payment"`

	Id                     string    `bun:"id,pk"`
	UserId                 string    `bun:"user_id"`
	Name                   string    `bun:"name"`
	CardNumber             string    `bun:"card_number"`
	CardExpirationYear     string    `bun:"card_expiration_year"`
	CardExpirationMonth    string    `bun:"card_expiration_month"`
	CardPassword           string    `bun:"card_password"`
	CustomerIdentityNumber string    `bun:"customer_identity_number"`
	BillingKey             string    `bun:"billing_key"`
	CreateTime             time.Time `bun:"create_time"`
	DeleteTime             time.Time `bun:"delete_time"`
}

func (u UserPayment) ToRedactedUserPayment() value.RedactedUserPayment {
	// TODO redact card number
	return value.RedactedUserPayment{
		Id:                             u.Id,
		UserId:                         u.UserId,
		Name:                           u.Name,
		RedactedCardNumber:             u.CardNumber,
		CardExpirationYear:             u.CardExpirationYear,
		CardExpirationMonth:            u.CardExpirationMonth,
		RedactedCardPassword:           "**",
		RedactedCustomerIdentityNumber: "**",
		CreateTime:                     u.CreateTime,
		DeleteTime:                     u.DeleteTime,
	}
}
