package request

// TODO(taekyeom) validation
type UserSignupRequest struct {
	FirstName               string `json:"firstName"`
	LastName                string `json:"lastName"`
	Gender                  string `json:"gender"`
	Birthday                string `json:"birthday"`
	Phone                   string `json:"phone"`
	AppOs                   string `json:"appOs"`
	AppVersion              string `json:"appVersion"`
	AppFcmToken             string `json:"appFcmToken"`
	SmsVerificationStateKey string `json:"smsVerificationStateKey"`
}

type UserUpdateRequest struct {
	Id          string `param:"userId"`
	AppOs       string `json:"appOs"`
	AppVersion  string `json:"appVersion"`
	AppFcmToken string `json:"appFcmToken"`
}

type DefaultPaymentUpdateRequest struct {
	Id        string `param:"userId"`
	PaymentId string `param:"paymentId"`
}
