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

type UserTaxiCallRequestResponse struct {
	Dryrun                    bool                   `json:"dryrun"`
	ToArrivalDistance         int                    `json:"toArrivalDistance"`
	ToArrivalETA              time.Duration          `json:"toArrivalEta"`
	ToArrivalPath             []value.Point          `json:"toArrivalPath"`
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
	CancelPenaltyPrice        int                    `json:"cancelPenaltyPrice"`
	TollFee                   int                    `json:"tollFee"`
	AdditionalPrice           int                    `json:"additionalPrice"`
	UsedPoint                 int                    `json:"used_point"`
	CurrentState              string                 `json:"currentState"`
	CreateTime                time.Time              `json:"createTime"`
	UpdateTime                time.Time              `json:"updateTime"`
}

type UserTaxiCallRequestPageResponse struct {
	PageToken string                        `json:"pageToken"`
	Data      []UserTaxiCallRequestResponse `json:"data"`
}

type DriverTaxiCallRequestPageResponse struct {
	PageToken string                          `json:"pageToken"`
	Data      []DriverTaxiCallRequestResponse `json:"data"`
}

func PaymentSummaryToResponse(paymentSummary value.PaymentSummary) PaymentSummaryResponse {
	return PaymentSummaryResponse{
		PaymentId:  paymentSummary.PaymentId,
		Company:    paymentSummary.Company,
		CardNumber: paymentSummary.CardNumber,
	}
}

type DriverTaxiCallRequestResponse struct {
	Dryrun                    bool           `json:"dryrun"`
	ToArrivalDistance         int            `json:"toArrivalDistance"`
	ToArrivalETA              time.Duration  `json:"toArrivalEta"`
	ToArrivalPath             []value.Point  `json:"toArrivalPath"`
	Id                        string         `json:"id"`
	UserId                    string         `json:"userId"`
	DriverId                  string         `json:"driverId"`
	Departure                 value.Location `json:"departure"`
	Arrival                   value.Location `json:"arrival"`
	Tags                      []string       `json:"tags"`
	UserTag                   string         `json:"userTag"`
	RequestBasePrice          int            `json:"requestBasePrice"`
	RequestMinAdditionalPrice int            `json:"requestMinAdditionalPrice"`
	RequestMaxAdditionalPrice int            `json:"requestMaxAdditionalPrice"`
	BasePrice                 int            `json:"basePrice"`
	CancelPenaltyPrice        int            `json:"cancelPenaltyPrice"`
	TollFee                   int            `json:"tollFee"`
	AdditionalPrice           int            `json:"additionalPrice"`
	// Deprecated: do not use this field
	DriverAdditionalRewardPrice int       `json:"driverAdditionalRewardPrice"`
	CurrentState                string    `json:"currentState"`
	CreateTime                  time.Time `json:"createTime"`
	UpdateTime                  time.Time `json:"updateTime"`
}

func UserTaxiCallRequestToResponse(taxiCallRequest entity.TaxiCallRequest) UserTaxiCallRequestResponse {
	resp := UserTaxiCallRequestResponse{
		Dryrun:            taxiCallRequest.Dryrun,
		ToArrivalDistance: taxiCallRequest.ToArrivalRoute.Distance,
		ToArrivalETA:      taxiCallRequest.ToArrivalRoute.ETA,
		ToArrivalPath:     append([]value.Point{}, taxiCallRequest.ToArrivalRoute.Path...),
		Id:                taxiCallRequest.Id,
		UserId:            taxiCallRequest.UserId,
		DriverId: func() string {
			if taxiCallRequest.DriverId.Valid {
				return taxiCallRequest.DriverId.String
			}
			return ""
		}(),
		Departure:                 taxiCallRequest.Departure,
		Arrival:                   taxiCallRequest.Arrival,
		Tags:                      append([]string{}, taxiCallRequest.Tags...),
		UserTag:                   taxiCallRequest.UserTag,
		Payment:                   PaymentSummaryToResponse(taxiCallRequest.PaymentSummary),
		RequestBasePrice:          taxiCallRequest.RequestBasePrice,
		RequestMinAdditionalPrice: taxiCallRequest.RequestMinAdditionalPrice,
		RequestMaxAdditionalPrice: taxiCallRequest.RequestMaxAdditionalPrice,
		BasePrice:                 taxiCallRequest.BasePrice,
		TollFee:                   taxiCallRequest.TollFee,
		AdditionalPrice:           taxiCallRequest.UserAdditionalPrice(),
		UsedPoint:                 taxiCallRequest.UserUsedPoint,
		CancelPenaltyPrice:        taxiCallRequest.CancelPenaltyPrice,
		CurrentState:              string(taxiCallRequest.CurrentState),
		CreateTime:                taxiCallRequest.CreateTime,
		UpdateTime:                taxiCallRequest.UpdateTime,
	}

	return resp
}

func DriverTaxiCallRequestToResponse(taxiCallRequest entity.TaxiCallRequest) DriverTaxiCallRequestResponse {
	resp := DriverTaxiCallRequestResponse{
		Dryrun:            taxiCallRequest.Dryrun,
		ToArrivalDistance: taxiCallRequest.ToArrivalRoute.Distance,
		ToArrivalETA:      taxiCallRequest.ToArrivalRoute.ETA,
		ToArrivalPath:     append([]value.Point{}, taxiCallRequest.ToArrivalRoute.Path...),
		Id:                taxiCallRequest.Id,
		UserId:            taxiCallRequest.UserId,
		DriverId: func() string {
			if taxiCallRequest.DriverId.Valid {
				return taxiCallRequest.DriverId.String
			}
			return ""
		}(),
		Departure:                 taxiCallRequest.Departure,
		Arrival:                   taxiCallRequest.Arrival,
		Tags:                      append([]string{}, taxiCallRequest.Tags...),
		UserTag:                   taxiCallRequest.UserTag,
		RequestBasePrice:          taxiCallRequest.RequestBasePrice,
		RequestMinAdditionalPrice: taxiCallRequest.RequestMinAdditionalPrice,
		RequestMaxAdditionalPrice: taxiCallRequest.RequestMaxAdditionalPrice,
		BasePrice:                 taxiCallRequest.BasePrice,
		CancelPenaltyPrice:        taxiCallRequest.DriverSettlementCancelPenaltyPrice(),
		TollFee:                   taxiCallRequest.TollFee,
		AdditionalPrice:           taxiCallRequest.DriverSettlementAdditonalPrice(),
		CurrentState:              string(taxiCallRequest.CurrentState),
		CreateTime:                taxiCallRequest.CreateTime,
		UpdateTime:                taxiCallRequest.UpdateTime,
	}

	return resp
}

type UserLatestTaxiCallRequestResponse struct {
	UserTaxiCallRequestResponse
	DriverPhone         string        `json:"driverPhone"`
	DriverCarNumber     string        `json:"driverCarNumber"`
	ToDepartureDistance int           `json:"toDepartureDistance"`
	ToDepartureETA      time.Duration `json:"toDepartureEta"`
	ToDeparturePath     []value.Point `json:"toDeparturePath"`
}

func UserLatestTaxiCallRequestToResponse(userLatestTaxiCallRequest entity.UserLatestTaxiCallRequest) UserLatestTaxiCallRequestResponse {
	return UserLatestTaxiCallRequestResponse{
		UserTaxiCallRequestResponse: UserTaxiCallRequestToResponse(userLatestTaxiCallRequest.TaxiCallRequest),
		DriverPhone:                 userLatestTaxiCallRequest.DriverPhone,
		DriverCarNumber:             userLatestTaxiCallRequest.DriverCarNumber,
		ToDepartureDistance:         userLatestTaxiCallRequest.ToDepartureRoute.Distance,
		ToDepartureETA:              userLatestTaxiCallRequest.ToDepartureRoute.ETA,
		ToDeparturePath:             append([]value.Point{}, userLatestTaxiCallRequest.ToDepartureRoute.Path...),
	}
}

type DriverLatestTaxiCallRequestResponse struct {
	DriverTaxiCallRequestResponse
	UserPhone           string        `json:"userPhone"`
	ToDepartureDistance int           `json:"toDepartureDistance"`
	ToDepartureETA      time.Duration `json:"toDepartureEta"`
	ToDeparturePath     []value.Point `json:"toDeparturePath"`
}

func DriverLatestTaxiCallRequestToResponse(driverLatestTaxiCallRequest entity.DriverLatestTaxiCallRequest) DriverLatestTaxiCallRequestResponse {
	return DriverLatestTaxiCallRequestResponse{
		DriverTaxiCallRequestResponse: DriverTaxiCallRequestToResponse(driverLatestTaxiCallRequest.TaxiCallRequest),
		UserPhone:                     driverLatestTaxiCallRequest.UserPhone,
		ToDepartureDistance:           driverLatestTaxiCallRequest.ToDepartureRoute.Distance,
		ToDepartureETA:                driverLatestTaxiCallRequest.ToDepartureRoute.ETA,
		ToDeparturePath:               append([]value.Point{}, driverLatestTaxiCallRequest.ToDepartureRoute.Path...),
	}
}

type DriverLatestTaxiCallRequestTicketResponse struct {
	DriverLatestTaxiCallRequestResponse
	TaxiCallTicketId string `json:"taxiCallTicketId"`
	Attempt          int    `json:"attempt"`
}

func DriverLatestTaxiCallRequestTicketToResponse(driverLatestTaxiCallRequestTicket entity.DriverLatestTaxiCallRequestTicket) DriverLatestTaxiCallRequestTicketResponse {
	return DriverLatestTaxiCallRequestTicketResponse{
		DriverLatestTaxiCallRequestResponse: DriverLatestTaxiCallRequestToResponse(driverLatestTaxiCallRequestTicket.DriverLatestTaxiCallRequest),
		TaxiCallTicketId:                    driverLatestTaxiCallRequestTicket.TicketId,
		Attempt:                             driverLatestTaxiCallRequestTicket.Attempt,
	}
}
