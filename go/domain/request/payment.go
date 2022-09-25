package request

type UserPaymentRegisterRequest struct {
	Name                   string `json:"name"`
	CardNumber             string `json:"cardNumber"`
	ExpirationYear         string `json:"expirationYear"`
	ExpirationMonth        string `json:"expirationMonth"`
	CustomerIdentityNumber string `json:"customerIdentityNumber"`
	DefaultPayment         bool   `json:"defaultPayment"`
}
