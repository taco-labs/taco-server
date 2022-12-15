package entity

import (
	"time"

	"github.com/taco-labs/taco/go/domain/value"
	"github.com/uptrace/bun"
)

const (
	MinimumPointToUse = 1000
	MinimumPriceToPay = 100
)

type UserPaymentRegistrationRequest struct {
	bun.BaseModel `bun:"table:user_payment_registration_request"`

	RequestId  int       `bun:"request_id,pk"`
	PaymentId  string    `bun:"payment_id"`
	UserId     string    `bun:"user_id"`
	CreateTime time.Time `bun:"create_time"`
}

type UserPayment struct {
	bun.BaseModel `bun:"table:user_payment"`

	Id                  string    `bun:"id,pk"`
	UserId              string    `bun:"user_id"`
	Name                string    `bun:"name"`
	CardCompany         string    `bun:"card_company"`
	RedactedCardNumber  string    `bun:"redacted_card_number"`
	BillingKey          string    `bun:"billing_key"`
	Invalid             bool      `bun:"invalid"`
	InvalidErrorCode    string    `bun:"invalid_error_code"`
	InvalidErrorMessage string    `bun:"invalid_error_message"`
	CreateTime          time.Time `bun:"create_time"`
	LastUseTime         time.Time `bun:"last_use_time"`
}

func (u UserPayment) ToSummary() value.PaymentSummary {
	return value.PaymentSummary{
		PaymentId:  u.Id,
		Company:    u.CardCompany,
		CardNumber: u.RedactedCardNumber,
	}
}

type UserPaymentTransactionRequest struct {
	bun.BaseModel `bun:"table:user_payment_transaction_request"`

	OrderId                    string               `bun:"order_id,pk"`
	UserId                     string               `bun:"user_id"`
	PaymentSummary             value.PaymentSummary `bun:"payment_summary"`
	OrderName                  string               `bun:"order_name"`
	Amount                     int                  `bun:"amount"`
	UsedPoint                  int                  `bun:"used_point"`
	SettlementAmount           int                  `bun:"settlement_amount"`
	AdditionalSettlementAmount int                  `bun:"additional_settlement_amount"`
	SettlementTargetId         string               `bun:"settlement_target_id"`
	Recovery                   bool                 `bun:"recovery"`
	CreateTime                 time.Time            `bun:"create_time"`
}

func (u UserPaymentTransactionRequest) GetPaymentAmount() int {
	return u.Amount - u.UsedPoint
}

func (u UserPaymentTransactionRequest) GetSettlementAmount() int {
	return u.SettlementAmount + u.AdditionalSettlementAmount
}

type UserPaymentOrder struct {
	bun.BaseModel `bun:"table:user_payment_order"`

	OrderId        string               `bun:"order_id,pk"`
	UserId         string               `bun:"user_id"`
	PaymentSummary value.PaymentSummary `bun:"payment_summary"`
	OrderName      string               `bun:"order_name"`
	Amount         int                  `bun:"amount"`
	UsedPoint      int                  `bun:"used_point"`
	PaymentKey     string               `bun:"payment_key"`
	ReceiptUrl     string               `bun:"receipt_url"`
	CreateTime     time.Time            `bun:"create_time"`
}

type UserPaymentFailedOrder struct {
	bun.BaseModel `bun:"table:user_payment_failed_order"`

	OrderId                    string    `bun:"order_id,pk"`
	UserId                     string    `bun:"user_id"`
	OrderName                  string    `bun:"order_name"`
	Amount                     int       `bun:"amount"`
	UsedPoint                  int       `bun:"used_point"`
	SettlementAmount           int       `bun:"settlement_amount"`
	AdditionalSettlementAmount int       `bun:"additional_settlement_amount"`
	SettlementTargetId         string    `bun:"settlement_target_id"`
	CreateTime                 time.Time `bun:"create_time"`
}

type UserPaymentPoint struct {
	bun.BaseModel `bun:"table:user_payment_point"`

	UserId string `bun:"user_id,pk"`
	Point  int    `bun:"point"`
}

func (u *UserPaymentPoint) UsePoint(price int) int {
	if u.Point < MinimumPointToUse {
		return 0
	}
	var pointToUse int
	if u.Point >= price {
		pointToUse = price
	} else if price < u.Point+MinimumPriceToPay {
		pointToUse = price - MinimumPriceToPay
	} else {
		pointToUse = u.Point
	}
	u.Point -= pointToUse

	return pointToUse
}

func (u *UserPaymentPoint) AddPoint(point int) {
	u.Point += point
}
