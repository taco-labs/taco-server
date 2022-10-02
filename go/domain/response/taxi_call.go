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
	Id                        string                 `json:"id"`
	UserId                    string                 `json:"userId"`
	DriverId                  *string                `json:"driverId"`
	Departure                 value.Location         `json:"departure"`
	Arrival                   value.Location         `json:"arrival"`
	Payment                   PaymentSummaryResponse `json:"payment"`
	RequestBasePrice          int                    `json:"requestBasePrice"`
	RequestMinAdditionalPrice int                    `json:"requestMinAdditionalPrice"`
	RequestMaxAdditionalPrice int                    `json:"requestMaxAdditionalPrice"`
	BasePrice                 int                    `json:"basePrice"`
	AdditionalPrice           int                    `json:"additionalPrice"`
	CurrentState              string                 `json:"currentState"`
	CreateTime                time.Time              `json:"createTime"`
	UpdateTime                time.Time              `json:"updateTime"`
}

type TaxiCallRequestHistoryResponse struct {
	TaxiCallState string    `json:"taxiCallState"`
	CreateTime    time.Time `json:"createTime"`
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
	return TaxiCallRequestResponse{
		Dryrun: taxiCallRequest.Dryrun,
		Id:     taxiCallRequest.Id,
		UserId: taxiCallRequest.UserId,
		DriverId: func() *string {
			if taxiCallRequest.DriverId.Valid {
				return &taxiCallRequest.DriverId.String
			}
			return nil
		}(),
		Departure:                 taxiCallRequest.Departure,
		Arrival:                   taxiCallRequest.Arrival,
		Payment:                   PaymentSummaryToResponse(taxiCallRequest.PaymentSummary),
		RequestBasePrice:          taxiCallRequest.RequestBasePrice,
		RequestMinAdditionalPrice: taxiCallRequest.RequestMinAdditionalPrice,
		RequestMaxAdditionalPrice: taxiCallRequest.RequestMaxAdditionalPrice,
		BasePrice:                 taxiCallRequest.BasePrice,
		AdditionalPrice:           taxiCallRequest.AdditionalPrice,
		CurrentState:              string(taxiCallRequest.CurrentState),
		CreateTime:                taxiCallRequest.CreateTime,
		UpdateTime:                taxiCallRequest.UpdateTime,
	}
}
