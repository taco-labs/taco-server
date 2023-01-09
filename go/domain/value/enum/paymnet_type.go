package enum

type PaymentType string

const (
	PaymentType_Card            PaymentType = "CARD"
	PaymentType_SignupPromition PaymentType = "SIGNUP_PROMOTION"
	PaymentType_Mock            PaymentType = "MOCK"
)
