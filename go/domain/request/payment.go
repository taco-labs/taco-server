package request

import "time"

type PaymentRegistrationCallbackRequest struct {
	RequestId   int    `json:"requestId"`
	BillingKey  string `json:"billingKey"`
	CardCompany string `json:"cardCompany"`
	CardNumber  string `json:"cardNumber"`
}

type PaymentTransactionSuccessCallbackRequest struct {
	OrderId    string
	PaymentKey string
	ReceiptUrl string
	CreateTime time.Time
}

type PaymentTransactionFailCallbackRequest struct {
	OrderId       string
	FailureCode   string
	FailureReason string
}
