package response

import (
	"net/url"
	"time"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/domain/value/enum"
	"github.com/taco-labs/taco/go/utils/slices"
)

type DriverResponse struct {
	Id                         string                `json:"id"`
	DriverType                 string                `json:"driverType"`
	FirstName                  string                `json:"firstName"`
	LastName                   string                `json:"lastName"`
	BirthDay                   string                `json:"birthday"`
	Phone                      string                `json:"phone"`
	Gender                     string                `json:"gender"`
	AppOs                      string                `json:"appOs"`
	AppVersion                 string                `json:"osVersion"`
	Active                     bool                  `json:"active"`
	OnDuty                     bool                  `json:"onDuty"`
	DriverLicenseId            string                `json:"driverLicenseId"`
	CompanyRegistrationNumber  string                `json:"companyRegistrationNumber"`
	CompanyName                string                `json:"companyName"`
	CarNumber                  string                `json:"carNumber"`
	ServiceRegion              string                `json:"serviceRegion"`
	DriverLicenseImageUploaded bool                  `json:"driverLicenseImageUploaded"`
	DriverProfileImageUploaded bool                  `json:"driverProfileImageUploaded"`
	UploadUrls                 value.DriverImageUrls `json:"uploadUrls"`
	DownloadUrls               value.DriverImageUrls `json:"downloadUrls"`
	ReferralCode               string                `json:"referralCode"`
}

type DriverSignupResponse struct {
	Token  string         `json:"token"`
	Driver DriverResponse `json:"driver"`
}

func DriverToResponse(driver entity.Driver) (DriverResponse, error) {
	referralCode, err := value.EncodeReferralCode(value.ReferralCode{
		ReferralType: enum.ReferralType_Driver,
		PhoneNumber:  driver.Phone,
	})
	if err != nil {
		return DriverResponse{}, err
	}
	return DriverResponse{
		Id:                         driver.Id,
		DriverType:                 string(driver.DriverType),
		FirstName:                  driver.FirstName,
		LastName:                   driver.LastName,
		BirthDay:                   driver.BirthDay,
		Phone:                      driver.Phone,
		Gender:                     driver.Gender,
		DriverLicenseId:            driver.DriverLicenseId,
		CarNumber:                  driver.CarNumber,
		ServiceRegion:              driver.ServiceRegion,
		CompanyRegistrationNumber:  driver.CompanyRegistrationNumber,
		CompanyName:                driver.CompanyName,
		AppOs:                      string(driver.AppOs),
		AppVersion:                 driver.AppVersion,
		Active:                     driver.Active,
		OnDuty:                     driver.OnDuty,
		DriverLicenseImageUploaded: driver.DriverLicenseImageUploaded,
		DriverProfileImageUploaded: driver.DriverProfileImageUploaded,
		UploadUrls:                 driver.UploadUrls,
		DownloadUrls:               driver.DownloadUrls,
		ReferralCode:               referralCode,
	}, nil
}

type DriverSettlemtnAccountResponse struct {
	DriverId      string    `json:"driverId"`
	Bank          string    `json:"bank"`
	AccountNumber string    `json:"accountNumber"`
	CreateTime    time.Time `json:"createTime"`
	UpdateTime    time.Time `json:"updateTime"`
}

func DriverSettlemtnAccountToResponse(account entity.DriverSettlementAccount) DriverSettlemtnAccountResponse {
	return DriverSettlemtnAccountResponse{
		DriverId:      account.DriverId,
		Bank:          account.Bank,
		AccountNumber: account.RedactedAccountNumber(),
		CreateTime:    account.CreateTime,
		UpdateTime:    account.UpdateTime,
	}
}

type DriverImageUrlResponse struct {
	UploadUrls   value.DriverImageUrls `json:"uploadUrls"`
	DownloadUrls value.DriverImageUrls `json:"downloadUrls"`
}

type DriverTotalSettlementResponse struct {
	DriverId                    string `json:"driverId"`
	TotalAmount                 int    `json:"totalAmount"`
	TotalAmountWithoutTax       int    `json:"TotalAmountWithoutTax"`
	RequestableAmount           int    `json:"requestableAmount"`
	RequestableAmountWithoutTax int    `json:"requestableAmountWithoutTax"`
}

func DriverTotalSettlementToResponse(driverTotalSettlement entity.DriverTotalSettlement) DriverTotalSettlementResponse {
	return DriverTotalSettlementResponse{
		DriverId:                    driverTotalSettlement.DriverId,
		TotalAmount:                 driverTotalSettlement.TotalAmount,
		TotalAmountWithoutTax:       entity.ExpectedSettlementAmountWithoutTax(driverTotalSettlement.TotalAmount),
		RequestableAmount:           driverTotalSettlement.RequestableAmount,
		RequestableAmountWithoutTax: entity.ExpectedSettlementAmountWithoutTax(driverTotalSettlement.RequestableAmount),
	}
}

type DriverSettlementHistoryResponse struct {
	DriverId         string    `json:"driverId"`
	Amount           int       `json:"amount"`
	AmountWithoutTax int       `json:"amountWithoutTax"`
	Bank             string    `json:"bank"`
	AccountNumber    string    `json:"accountNumber"`
	RequestTime      time.Time `json:"requestTime"`
	CreateTime       time.Time `json:"createTime"`
}

func DriverSettlementHistoryToResponse(settlementHistory entity.DriverSettlementHistory) DriverSettlementHistoryResponse {
	return DriverSettlementHistoryResponse{
		DriverId:         settlementHistory.DriverId,
		Amount:           settlementHistory.Amount,
		AmountWithoutTax: settlementHistory.AmountWithoutTax,
		Bank:             settlementHistory.Bank,
		AccountNumber:    settlementHistory.RedactedAccountNumber(),
		RequestTime:      settlementHistory.RequestTime,
		CreateTime:       settlementHistory.CreateTime,
	}
}

type ListDriverSettlementHistoryResponse struct {
	PageToken string                            `json:"pageToken"`
	Histories []DriverSettlementHistoryResponse `json:"histories"`
}

func ListDriverSettlementHistoryToResponse(settlementHistories []entity.DriverSettlementHistory, pageToken time.Time) ListDriverSettlementHistoryResponse {
	return ListDriverSettlementHistoryResponse{
		PageToken: url.QueryEscape(pageToken.Format(time.RFC3339Nano)),
		Histories: slices.Map(settlementHistories, DriverSettlementHistoryToResponse),
	}
}

type DriverSettlementTransferResponse struct {
	ExpectedTransferAmount int `json:"expectedTransferAmount"`
}
