package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/domain/value/enum"
	"github.com/uptrace/bun"
)

type Driver struct {
	DriverDto
	DownloadUrls value.DriverImageUrls
	UploadUrls   value.DriverImageUrls
}

type DriverDto struct {
	bun.BaseModel `bun:"table:driver"`

	Id                         string          `bun:"id,pk"`
	DriverType                 enum.DriverType `bun:"driver_type"`
	FirstName                  string          `bun:"first_name"`
	LastName                   string          `bun:"last_name"`
	BirthDay                   string          `bun:"birthday"`
	Phone                      string          `bun:"phone"`
	Gender                     string          `bun:"gender"`
	AppOs                      enum.OsType     `bun:"app_os"`
	AppVersion                 string          `bun:"app_version"`
	UserUniqueKey              string          `bun:"user_unique_key"`
	DriverLicenseId            string          `bun:"driver_license_id"`
	CompanyRegistrationNumber  string          `bun:"company_registration_number"`
	CompanyName                string          `bun:"company_name"`
	ServiceRegion              string          `bun:"service_region"`
	CarNumber                  string          `bun:"car_number"`
	DriverLicenseImageUploaded bool            `bun:"driver_license_image_uploaded"`
	DriverProfileImageUploaded bool            `bun:"driver_profile_image_uploaded"`
	Active                     bool            `bun:"active"`
	OnDuty                     bool            `bun:"on_duty"`
	CreateTime                 time.Time       `bun:"create_time"`
	UpdateTime                 time.Time       `bun:"update_time"`
	DeleteTime                 time.Time       `bun:"delete_time"`
}

func (d DriverDto) MockAccount() bool {
	return d.Id == uuid.Nil.String()
}

func (d DriverDto) FullName() string {
	return fmt.Sprintf("%s%s", d.LastName, d.FirstName)
}

func (d DriverDto) ReferralCode() string {
	referralCode, _ := value.EncodeReferralCode(value.ReferralCode{
		ReferralType: enum.ReferralType_Driver,
		PhoneNumber:  d.Phone,
	})

	return referralCode
}

type DriverSettlementAccount struct {
	bun.BaseModel `bun:"table:driver_settlement_account"`

	DriverId          string    `bun:"driver_id,pk"`
	Bank              string    `bun:"bank"`
	AccountNumber     string    `bun:"account_number"`
	BankTransactionId string    `bun:"bank_transaction_id"`
	CreateTime        time.Time `bun:"create_time"`
	UpdateTime        time.Time `bun:"update_time"`
}

func (d DriverSettlementAccount) RedactedAccountNumber() string {
	lastAccountNumber := d.AccountNumber[len(d.AccountNumber)-4:]
	return fmt.Sprintf("****%s", lastAccountNumber)
}

type DriverLocation struct {
	Location   value.Point
	DriverId   string
	UpdateTime time.Time
}

type DriverResidentRegistrationNumber struct {
	bun.BaseModel `bun:"table:driver_resident_registration_number"`

	DriverId                            string `bun:"driver_id,pk"`
	EncryptedResidentRegistrationNumber []byte `bun:"encrypted_resident_registration_number"`
}
