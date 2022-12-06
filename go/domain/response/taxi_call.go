package response

import (
	"time"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
)

type PaymentSummaryResponse struct {
	PaymentId  string `json:"paymentId"`
	Company    string `json:"company"`
	CardNumber string `json:"cardNumber"`
}

type TaxiCallRequestResponse struct {
	Dryrun                    bool                   `json:"dryrun"`
	Distance                  int                    `json:"distance"`
	ETA                       time.Duration          `json:"eta"`
	Path                      []value.Point          `json:"path"`
	Id                        string                 `json:"id"`
	UserId                    string                 `json:"userId"`
	DriverId                  string                 `json:"driverId"`
	Departure                 value.Location         `json:"departure"`
	Arrival                   value.Location         `json:"arrival"`
	Tags                      []string               `json:"tags"`
	UserTag                   string                 `json:"userTag"`
	Payment                   PaymentSummaryResponse `json:"payment"`
	RequestBasePrice          int                    `json:"requestBasePrice"`
	RequestMinAdditionalPrice int                    `json:"requestMinAdditionalPrice"`
	RequestMaxAdditionalPrice int                    `json:"requestMaxAdditionalPrice"`
	BasePrice                 int                    `json:"basePrice"`
	TollFee                   int                    `json:"tollFee"`
	AdditionalPrice           int                    `json:"additionalPrice"`
	CurrentState              string                 `json:"currentState"`
	CreateTime                time.Time              `json:"createTime"`
	UpdateTime                time.Time              `json:"updateTime"`
}

type TaxiCallRequestPageResponse struct {
	PageToken string                    `json:"pageToken"`
	Data      []TaxiCallRequestResponse `json:"data"`
}

func PaymentSummaryToResponse(paymentSummary value.PaymentSummary) PaymentSummaryResponse {
	return PaymentSummaryResponse{
		PaymentId:  paymentSummary.PaymentId,
		Company:    paymentSummary.Company,
		CardNumber: paymentSummary.CardNumber,
	}
}

func TaxiCallRequestToResponse(taxiCallRequest entity.TaxiCallRequest) TaxiCallRequestResponse {
	resp := TaxiCallRequestResponse{
		Dryrun:   taxiCallRequest.Dryrun,
		Distance: taxiCallRequest.Route.Distance,
		ETA:      taxiCallRequest.Route.ETA,
		Path:     taxiCallRequest.Route.Path,
		Id:       taxiCallRequest.Id,
		UserId:   taxiCallRequest.UserId,
		DriverId: func() string {
			if taxiCallRequest.DriverId.Valid {
				return taxiCallRequest.DriverId.String
			}
			return ""
		}(),
		Departure:                 taxiCallRequest.Departure,
		Arrival:                   taxiCallRequest.Arrival,
		Tags:                      taxiCallRequest.Tags,
		UserTag:                   taxiCallRequest.UserTag,
		Payment:                   PaymentSummaryToResponse(taxiCallRequest.PaymentSummary),
		RequestBasePrice:          taxiCallRequest.RequestBasePrice,
		RequestMinAdditionalPrice: taxiCallRequest.RequestMinAdditionalPrice,
		RequestMaxAdditionalPrice: taxiCallRequest.RequestMaxAdditionalPrice,
		BasePrice:                 taxiCallRequest.BasePrice,
		TollFee:                   taxiCallRequest.TollFee,
		AdditionalPrice:           taxiCallRequest.AdditionalPrice,
		CurrentState:              string(taxiCallRequest.CurrentState),
		CreateTime:                taxiCallRequest.CreateTime,
		UpdateTime:                taxiCallRequest.UpdateTime,
	}

	// TODO (taekyeom) sanitization must be performed in entity area..
	if resp.Tags == nil {
		resp.Tags = []string{}
	}
	if resp.Path == nil {
		resp.Path = []value.Point{}
	}

	return resp
}

type UserLatestTaxiCallRequestResponse struct {
	TaxiCallRequestResponse
	DriverPhone     string `json:"driverPhone"`
	DriverCarNumber string `json:"driverCarNumber"`
}

func UserLatestTaxiCallRequestToResponse(userLatestTaxiCallRequest entity.UserLatestTaxiCallRequest) UserLatestTaxiCallRequestResponse {
	return UserLatestTaxiCallRequestResponse{
		TaxiCallRequestResponse: TaxiCallRequestToResponse(userLatestTaxiCallRequest.TaxiCallRequest),
		DriverPhone:             userLatestTaxiCallRequest.DriverPhone,
		DriverCarNumber:         userLatestTaxiCallRequest.DriverCarNumber,
	}
}

type DriverLatestTaxiCallRequestResponse struct {
	TaxiCallRequestResponse
	UserPhone string `json:"userPhone"`
}

func DriverLatestTaxiCallRequestToResponse(driverLatestTaxiCallRequest entity.DriverLatestTaxiCallRequest) DriverLatestTaxiCallRequestResponse {
	return DriverLatestTaxiCallRequestResponse{
		TaxiCallRequestResponse: TaxiCallRequestToResponse(driverLatestTaxiCallRequest.TaxiCallRequest),
		UserPhone:               driverLatestTaxiCallRequest.UserPhone,
	}
}
