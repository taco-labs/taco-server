package request

type PaymentRegistrationCallbackRequest struct {
	RequestId   int    `json:"requestId"`
	BillingKey  string `json:"billingKey"`
	CardCompany string `json:"cardCompany"`
	CardNumber  string `json:"cardNumber"`
}
