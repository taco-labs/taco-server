package request

type SmsVerificationRequest struct {
	StateKey string `json:"stateKey"`
	Phone    string `json:"phone"`
}

type SmsSigninRequest struct {
	StateKey         string `json:"stateKey"`
	VerificationCode string `json:"verificationCode"`
}
