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
	CompanyRegistrationNumber string `json:"companyRegistrationNumber"`
	CompanyName               string `json:"companyName"`
	// Deprecated: use driver car profile apis
	TaxiCategory string `json:"taxiCategory"`
	// Deprecated: use driver car profile apis
	CarNumber string `json:"carNumber"`
	// Deprecated: use driver car profile apis
	CarModel                   string `json:"carModel"`
	ServiceRegion              string `json:"serviceRegion"`
	ResidentRegistrationNumber string `json:"residentRegistrationNumber"`
	ReferralCode               string `json:"referralCode"`
}

type DriverUpdateRequest struct {
	Id                   string `param:"driverId"`
	AppOs                string `json:"appOs"`
	AppVersion           string `json:"appVersion"`
	AppFcmToken          string `json:"appFcmToken"`
	LicenseImageUploaded bool   `json:"licenseImageUploaded"`
	ProfileImageUploaded bool   `json:"profileImageUploaded"`
	// Deprecated: use car profile instead
	CarNumber string `json:"carNumber"`
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

type ListNonActivatedDriverRequest struct {
	PageToken string `param:"pageToken"`
	Count     int    `param:"count"`
}

type AddCarProfileRequest struct {
	SelectAsProfile bool   `json:"selectAsProfile"`
	TaxiCategory    string `json:"taxiCategory"`
	CarNumber       string `json:"carNumber"`
	CarModel        string `json:"carModel"`
}

type UpdateCarProfileRequest struct {
	TaxiCategory string `json:"taxiCategory"`
	ProfileId    string `param:"carProfileId"`
	CarNumber    string `json:"carNumber"`
	CarModel     string `json:"carModel"`
}
