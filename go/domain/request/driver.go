package request

// TODO(taekyeom) validation
type DriverSignupRequest struct {
	DriverType      string `json:"driverType"`
	FirstName       string `json:"firstName"`
	LastName        string `json:"lastName"`
	IamUid          string `json:"iamUid"`
	AppOs           string `json:"appOs"`
	AppVersion      string `json:"appVersion"`
	AppFcmToken     string `json:"appFcmToken"`
	DriverLicenseId string `json:"driverLicenseId"`
}

type DriverOnDutyUpdateRequest struct {
	DriverId string `param:"driverId"`
	OnDuty   bool   `json:"onDuty"`
}

type DriverLocationUpdateRequest struct {
	DriverId  string  `param:"driverId"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type MockDriverIdentity struct {
	MockGender   string `json:"gender"`
	MockBirthday string `json:"birthday"`
	MockPhone    string `json:"phone"`
}
