package request

// TODO(taekyeom) validation
type UserSignupRequest struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Email       string `json:"email"`
	IamUid      string `json:"iamUid"`
	AppOs       string `json:"appOs"`
	OsVersion   string `json:"osVersion"`
	AppVersion  string `json:"appVersion"`
	AppFcmToken string `json:"appFcmToken"`
}

type UserUpdateRequest struct {
	Id          string `param:"userId"`
	AppFcmToken string `json:"appFcmToken"`
}

type DefaultPaymentUpdateRequest struct {
	Id               string `param:"userId"`
	DefaultPaymentId string `json:"defaultPaymentId"`
}

// For mock user identity test
type MockUserIdentity struct {
	MockGender   string `json:"gender"`
	MockBirthday string `json:"birthday"`
	MockPhone    string `json:"phone"`
}
