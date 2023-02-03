package response

import (
	"time"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
)

type UserTaxiCallRequestHistoryResponse struct {
	Id                 string                 `json:"id"`
	DriverId           string                 `json:"driverId"`
	Departure          value.Location         `json:"departure"`
	Arrival            value.Location         `json:"arrival"`
	Tags               []string               `json:"tags"`
	UserTag            string                 `json:"userTag"`
	Payment            PaymentSummaryResponse `json:"payment"`
	BasePrice          int                    `json:"basePrice"`
	CancelPenaltyPrice int                    `json:"cancelPenaltyPrice"`
	TollFee            int                    `json:"tollFee"`
	AdditionalPrice    int                    `json:"additionalPrice"`
	// Deprecated. use UsedPoint2 instead.
	UsedPoint    int       `json:"used_point"`
	UsedPoint2   int       `json:"usedPoint"`
	CurrentState string    `json:"currentState"`
	CreateTime   time.Time `json:"createTime"`
	UpdateTime   time.Time `json:"updateTime"`
}

func UserTaxiCallRequestToHistoryResponse(taxiCallRequest entity.TaxiCallRequest) UserTaxiCallRequestHistoryResponse {
	return UserTaxiCallRequestHistoryResponse{
		Id: taxiCallRequest.Id,
		DriverId: func() string {
			if taxiCallRequest.DriverId.Valid {
				return taxiCallRequest.DriverId.String
			}
			return ""
		}(),
		Departure:          taxiCallRequest.Departure,
		Arrival:            taxiCallRequest.Arrival,
		Tags:               append([]string{}, taxiCallRequest.Tags...),
		UserTag:            taxiCallRequest.UserTag,
		Payment:            PaymentSummaryToResponse(taxiCallRequest.PaymentSummary),
		BasePrice:          taxiCallRequest.BasePrice,
		TollFee:            taxiCallRequest.TollFee,
		AdditionalPrice:    taxiCallRequest.UserAdditionalPrice(),
		UsedPoint:          taxiCallRequest.UserUsedPoint,
		UsedPoint2:         taxiCallRequest.UserUsedPoint,
		CancelPenaltyPrice: taxiCallRequest.CancelPenaltyPrice,
		CurrentState:       string(taxiCallRequest.CurrentState),
		CreateTime:         taxiCallRequest.CreateTime,
		UpdateTime:         taxiCallRequest.UpdateTime,
	}
}

type UserTaxiCallRequestHistoryPageResponse struct {
	PageToken string                               `json:"pageToken"`
	Data      []UserTaxiCallRequestHistoryResponse `json:"data"`
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
	// Deprecated. use UsedPoint2 instead.
	UsedPoint    int       `json:"used_point"`
	UsedPoint2   int       `json:"usedPoint"`
	CurrentState string    `json:"currentState"`
	CreateTime   time.Time `json:"createTime"`
	UpdateTime   time.Time `json:"updateTime"`
}

func UserTaxiCallRequestToResponse(taxiCallRequest entity.TaxiCallRequest) UserTaxiCallRequestResponse {
	resp := UserTaxiCallRequestResponse{
		Dryrun:            taxiCallRequest.Dryrun,
		ToArrivalDistance: taxiCallRequest.ToArrivalRoute.Route.Distance,
		ToArrivalETA:      taxiCallRequest.ToArrivalRoute.Route.ETA,
		ToArrivalPath:     append([]value.Point{}, taxiCallRequest.ToArrivalRoute.Route.Path...),
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
		UsedPoint2:                taxiCallRequest.UserUsedPoint,
		CancelPenaltyPrice:        taxiCallRequest.CancelPenaltyPrice,
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
		ToDepartureDistance:         userLatestTaxiCallRequest.ToDepartureRoute.Route.Distance,
		ToDepartureETA:              userLatestTaxiCallRequest.ToDepartureRoute.Route.ETA,
		ToDeparturePath:             append([]value.Point{}, userLatestTaxiCallRequest.ToDepartureRoute.Route.Path...),
	}
}

type UserLatestTaxiCallRequestTicketResponse struct {
	TaxiCallRequestId    string    `json:"taxiCallRequestId"`
	AdditionalPrice      int       `json:"additionalPrice"`
	Attempt              int       `json:"attempt"`
	SearchRangeInMinutes int       `json:"SearchRangeInMinutes"`
	SearchRangeInMeters  int       `json:"SearchRangeInMeters"`
	UpdateTime           time.Time `json:"updateTime"`
}

func TaxiCallTicketToResponse(ticket entity.TaxiCallTicket) UserLatestTaxiCallRequestTicketResponse {
	return UserLatestTaxiCallRequestTicketResponse{
		TaxiCallRequestId:    ticket.TaxiCallRequestId,
		AdditionalPrice:      ticket.UserAdditionalPrice(),
		Attempt:              ticket.AdditionalPrice,
		SearchRangeInMinutes: ticket.GetRadiusMinutes(),
		SearchRangeInMeters:  ticket.GetRadius(),
		UpdateTime:           ticket.CreateTime,
	}
}
