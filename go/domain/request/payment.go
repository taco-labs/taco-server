package request

type UserPaymentRegisterRequest struct {
	Name                   string `json:"name"`
	CardNumber             string `json:"cardNumber"`
	CardExpirationYear     string `json:"cardExpirationYear"`
	CardExpirationMonth    string `json:"cardExpirationMonth"`
	CardPassword           string `json:"cardPassword"`
	CustomerIdentityNumber string `json:"customerIdentityNumber"`
	DefaultPayment         bool   `json:"defaultPayment"`
}
