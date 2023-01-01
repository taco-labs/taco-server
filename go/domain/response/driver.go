package response

import (
	"net/url"
	"time"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/utils/slices"
)

type DriverDtoResponse struct {
	Id                         string `json:"id"`
	DriverType                 string `json:"driverType"`
	FirstName                  string `json:"firstName"`
	LastName                   string `json:"lastName"`
	BirthDay                   string `json:"birthday"`
	Phone                      string `json:"phone"`
	Gender                     string `json:"gender"`
	AppOs                      string `json:"appOs"`
	AppVersion                 string `json:"appVersion"`
	Active                     bool   `json:"active"`
	OnDuty                     bool   `json:"onDuty"`
	DriverLicenseId            string `json:"driverLicenseId"`
	CompanyRegistrationNumber  string `json:"companyRegistrationNumber"`
	CompanyName                string `json:"companyName"`
	CarNumber                  string `json:"carNumber"`
	ServiceRegion              string `json:"serviceRegion"`
	DriverLicenseImageUploaded bool   `json:"driverLicenseImageUploaded"`
	DriverProfileImageUploaded bool   `json:"driverProfileImageUploaded"`
	ReferralCode               string `json:"referralCode"`
}

func DriverDtoToResponse(driverDto entity.DriverDto) DriverDtoResponse {
	return DriverDtoResponse{
		Id:                         driverDto.Id,
		DriverType:                 string(driverDto.DriverType),
		FirstName:                  driverDto.FirstName,
		LastName:                   driverDto.LastName,
		BirthDay:                   driverDto.BirthDay,
		Phone:                      driverDto.Phone,
		Gender:                     driverDto.Gender,
		DriverLicenseId:            driverDto.DriverLicenseId,
		CarNumber:                  driverDto.CarNumber,
		ServiceRegion:              driverDto.ServiceRegion,
		CompanyRegistrationNumber:  driverDto.CompanyRegistrationNumber,
		CompanyName:                driverDto.CompanyName,
		AppOs:                      string(driverDto.AppOs),
		AppVersion:                 driverDto.AppVersion,
		Active:                     driverDto.Active,
		OnDuty:                     driverDto.OnDuty,
		DriverLicenseImageUploaded: driverDto.DriverLicenseImageUploaded,
		DriverProfileImageUploaded: driverDto.DriverProfileImageUploaded,
		ReferralCode:               driverDto.ReferralCode(),
	}
}

type DriverResponse struct {
	DriverDtoResponse
	UploadUrls   value.DriverImageUrls `json:"uploadUrls"`
	DownloadUrls value.DriverImageUrls `json:"downloadUrls"`
}

type DriverSignupResponse struct {
	Token  string         `json:"token"`
	Driver DriverResponse `json:"driver"`
}

func DriverToResponse(driver entity.Driver) DriverResponse {
	return DriverResponse{
		DriverDtoResponse: DriverDtoToResponse(driver.DriverDto),
		UploadUrls:        driver.UploadUrls,
		DownloadUrls:      driver.DownloadUrls,
	}
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

type ListNonActivatedDriverResponse struct {
	PageToken string              `json:"pageToken"`
	Drivers   []DriverDtoResponse `json:"drivers"`
}
