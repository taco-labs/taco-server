package response

import (
	"time"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
)

type DriverTaxiCallRequestHistoryResponse struct {
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
	CurrentState       string                 `json:"currentState"`
	CreateTime         time.Time              `json:"createTime"`
	UpdateTime         time.Time              `json:"updateTime"`
}

func DriverTaxiCallRequestToHistoryResponse(taxiCallRequest entity.TaxiCallRequest) DriverTaxiCallRequestHistoryResponse {
	return DriverTaxiCallRequestHistoryResponse{
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
		AdditionalPrice:    taxiCallRequest.DriverSettlementAdditonalPrice(),
		CancelPenaltyPrice: taxiCallRequest.DriverSettlementCancelPenaltyPrice(),
		CurrentState:       string(taxiCallRequest.CurrentState),
		CreateTime:         taxiCallRequest.CreateTime,
		UpdateTime:         taxiCallRequest.UpdateTime,
	}
}

type DriverTaxiCallRequestHistoryPageResponse struct {
	PageToken string                                 `json:"pageToken"`
	Data      []DriverTaxiCallRequestHistoryResponse `json:"data"`
}

func PaymentSummaryToResponse(paymentSummary value.PaymentSummary) PaymentSummaryResponse {
	return PaymentSummaryResponse{
		PaymentId:   paymentSummary.PaymentId,
		Company:     paymentSummary.Company,
		CardNumber:  paymentSummary.CardNumber,
		LastUseTime: paymentSummary.LastUseTime,
	}
}

type DriverTaxiCallRequestResponse struct {
	Dryrun            bool          `json:"dryrun"`
	ToArrivalDistance int           `json:"toArrivalDistance"`
	ToArrivalETA      time.Duration `json:"toArrivalEta"`
	// Deprecated. not used in client, delete it later.
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

func DriverTaxiCallRequestToResponse(taxiCallRequest entity.TaxiCallRequest) DriverTaxiCallRequestResponse {
	resp := DriverTaxiCallRequestResponse{
		Dryrun:            taxiCallRequest.Dryrun,
		ToArrivalDistance: taxiCallRequest.ToArrivalRoute.Route.Distance,
		ToArrivalETA:      taxiCallRequest.ToArrivalRoute.Route.ETA,
		ToArrivalPath:     []value.Point{},
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

type DriverLatestTaxiCallRequestResponse struct {
	DriverTaxiCallRequestResponse
	UserPhone           string `json:"userPhone"`
	ToDepartureDistance int    `json:"toDepartureDistance"`
	// Deprecated. not used in client, delete it later.
	ToDepartureETA time.Duration `json:"toDepartureEta"`
	// Deprecated. not used in client, delete it later.
	ToDeparturePath []value.Point `json:"toDeparturePath"`
}

func DriverLatestTaxiCallRequestToResponse(driverLatestTaxiCallRequest entity.DriverLatestTaxiCallRequest) DriverLatestTaxiCallRequestResponse {
	return DriverLatestTaxiCallRequestResponse{
		DriverTaxiCallRequestResponse: DriverTaxiCallRequestToResponse(driverLatestTaxiCallRequest.TaxiCallRequest),
		UserPhone:                     driverLatestTaxiCallRequest.UserPhone,
		ToDepartureDistance:           driverLatestTaxiCallRequest.ToDepartureRoute.Route.Distance,
		ToDeparturePath:               []value.Point{},
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

type DriverTaxiCallContextResponse struct {
	DriverId            string `json:"driverId"`
	CanReceive          bool   `json:"canReceive"`
	ToDepartureDistance int    `json:"toDepartureDistance"` // In meter
}

func DriverTaxiCallContextToResponse(d entity.DriverTaxiCallContext) DriverTaxiCallContextResponse {
	return DriverTaxiCallContextResponse{
		DriverId:            d.DriverId,
		CanReceive:          d.CanReceive,
		ToDepartureDistance: d.ToDepartureDistance,
	}
}
