package value

import (
	"time"

	"github.com/taco-labs/taco/go/domain/value/enum"
)

type PaymentRegistrationRequestParam struct {
	RequestId       int
	UserPhone       string
	AuthKey         string
	RegistrationUrl string
}

type Payment struct {
	OrderId   string
	Amount    int
	OrderName string
}

type PaymentResult struct {
	OrderId    string
	PaymentKey string
	Amount     int
	OrderName  string
	ReceiptUrl string
}

type PaymentCancel struct {
	PaymentKey   string
	CancelAmount int
	Reason       string
}

type PaymentSummary struct {
	PaymentId   string           `json:"paymentId"`
	PaymentType enum.PaymentType `json:"paymentType"`
	Company     string           `json:"company"`
	CardNumber  string           `json:"cardNumber"`
	LastUseTime time.Time        `json:"lastUseTime"`
}

type CardPaymentInfo struct {
	CustomerKey         string
	CardCompany         string
	CardNumber          string
	CardExpirationYear  string
	CardExpirationMonth string
	BillingKey          string
}
