package request

// TODO(taekyeom) validation
type DriverSignupRequest struct {
	DriverType                string `json:"driverType"`
	FirstName                 string `json:"firstName"`
	LastName                  string `json:"lastName"`
	Gender                    string `json:"gender"`
	Birthday                  string `json:"birthday"`
	Phone                     string `json:"phone"`
	AppOs                     string `json:"appOs"`
	AppVersion                string `json:"appVersion"`
	AppFcmToken               string `json:"appFcmToken"`
	DriverLicenseId           string `json:"driverLicenseId"`
	SmsVerificationStateKey   string `json:"smsVerificationStateKey"`
	CompanyRegistrationNumber string `json:"CompanyRegistrationNumber"`
	CompanyName               string `json:"companyName"`
	CarNumber                 string `json:"CarNumber"`
	ServiceRegion             string `json:"serviceRegion"`
}

type DriverUpdateRequest struct {
	Id                   string `param:"driverId"`
	AppOs                string `json:"appOs"`
	AppVersion           string `json:"appVersion"`
	AppFcmToken          string `json:"appFcmToken"`
	LicenseImageUploaded bool   `json:"licenseImageUploaded"`
	ProfileImageUploaded bool   `json:"profileImageUploaded"`
	CarNumber            string `json:"carNumber"`
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

type DriverSettlementAccountRegisterRequest struct {
	DriverId      string `param:"driverId"`
	Bank          string `json:"bank"`
	AccountNumber string `json:"accountNumber"`
}

type DriverSettlementAccountUpdateRequest struct {
	DriverId      string `param:"driverId"`
	Bank          string `json:"bank"`
	AccountNumber string `json:"accountNumber"`
}
