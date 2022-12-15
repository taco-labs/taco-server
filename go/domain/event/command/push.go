package command

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/utils"
)

var (
	EventUri_PushPrefix                 = "Push/"
	EventUri_RawMessage                 = fmt.Sprintf("%sRawMessage", EventUri_PushPrefix)
	EventUri_UserTaxiCallNotification   = fmt.Sprintf("%sUserTaxiCall", EventUri_PushPrefix)
	EventUri_DriverTaxiCallNotification = fmt.Sprintf("%sDriverTaxiCall", EventUri_PushPrefix)
)

type PushUserTaxiCallCommand struct {
	UserId               string         `json:"userId"`
	TaxiCallRequestId    string         `json:"taxiCallRequestId"`
	TaxiCallState        string         `json:"taxiCallState"`
	RequestBasePrice     int            `json:"requestBasePrice,omitempty"`
	AdditionalPrice      int            `json:"additionalPrice,omitempty"`
	UsedPoint            int            `json:"usedPoint"`
	DriverId             string         `json:"driverId,omitempty"`
	BasePrice            int            `json:"basePrice,omitempty"`
	DriverLocation       value.Point    `json:"driverLocation,omitempty"`
	Departure            value.Location `json:"departureAddress,omitempty"`
	Arrival              value.Location `json:"arrivalAddress,omitempty"`
	ToDepartureRoute     value.Route    `json:"toDepartureRoute"`
	ToArrivalRoute       value.Route    `json:"toArrivalRoute"`
	SearchRangeInMinutes int            `json:"searchRangeInMinutes,omitempty"`
	UpdateTime           time.Time      `json:"updateTime"`
}

type PushDriverTaxiCallCommand struct {
	DriverId                    string         `json:"driverId"`
	UserId                      string         `json:"userId"`
	TaxiCallRequestId           string         `json:"taxiCallRequestId"`
	TaxiCallTicketId            string         `json:"taxiCallTicketId"`
	TaxiCallState               string         `json:"taxiCallState"`
	DriverLocation              value.Point    `json:"driverLocation"`
	RequestBasePrice            int            `json:"requestBasePrice,omitempty"`
	AdditionalPrice             int            `json:"additionalPrice,omitempty"`
	DriverAdditionalRewardPrice int            `json:"driverAdditionalRewardPrice"`
	Departure                   value.Location `json:"departureAddress,omitempty"`
	Arrival                     value.Location `json:"arrivalAddress,omitempty"`
	ToArrivalRoute              value.Route    `json:"toArrivalRoute"`
	ToDepartureDistance         int            `json:"toDepartureDistance"`
	Tags                        []string       `json:"tags"`
	UserTag                     string         `json:"userTag"`
	Attempt                     int            `json:"attempt"`
	UpdateTime                  time.Time      `json:"updateTime"`
}

type PushRawCommand struct {
	AccountId    string            `json:"accountId"`
	Category     string            `json:"category"`
	MessageTitle string            `json:"messageTitle"`
	MessageBody  string            `json:"messageBody"`
	Data         map[string]string `json:"data"`
}

func NewPushUserTaxiCallCommand(taxiCallRequest entity.TaxiCallRequest,
	taxiCallTicket entity.TaxiCallTicket,
	driverTaxiCallContext entity.DriverTaxiCallContext,
	updateTime time.Time,
) entity.Event {
	cmd := PushUserTaxiCallCommand{
		UserId:               taxiCallRequest.UserId,
		TaxiCallRequestId:    taxiCallRequest.Id,
		TaxiCallState:        string(taxiCallRequest.CurrentState),
		RequestBasePrice:     taxiCallRequest.RequestBasePrice,
		AdditionalPrice:      taxiCallTicket.UserAdditionalPrice(),
		UsedPoint:            taxiCallRequest.UserUsedPoint,
		DriverId:             taxiCallRequest.DriverId.String,
		BasePrice:            taxiCallRequest.BasePrice,
		DriverLocation:       driverTaxiCallContext.Location,
		Departure:            taxiCallRequest.Departure,
		Arrival:              taxiCallRequest.Arrival,
		ToDepartureRoute:     taxiCallRequest.ToDepartureRoute,
		ToArrivalRoute:       taxiCallRequest.ToArrivalRoute,
		SearchRangeInMinutes: taxiCallTicket.GetRadiusMinutes(),
		UpdateTime:           updateTime,
	}

	cmdJson, _ := json.Marshal(cmd)

	return entity.Event{
		MessageId:    utils.MustNewUUID(),
		EventUri:     EventUri_UserTaxiCallNotification,
		DelaySeconds: 0,
		Payload:      cmdJson,
	}
}

func NewPushDriverTaxiCallCommand(
	driverId string,
	taxiCallRequest entity.TaxiCallRequest,
	taxiCallTicket entity.TaxiCallTicket,
	driverTaxiCallContext entity.DriverTaxiCallContext,
	updateTime time.Time,
) entity.Event {

	cmd := PushDriverTaxiCallCommand{
		DriverId:                    driverId,
		Attempt:                     taxiCallTicket.Attempt,
		UserId:                      taxiCallRequest.UserId,
		TaxiCallRequestId:           taxiCallRequest.Id,
		TaxiCallTicketId:            taxiCallTicket.TicketId,
		TaxiCallState:               string(taxiCallRequest.CurrentState),
		DriverLocation:              driverTaxiCallContext.Location,
		RequestBasePrice:            taxiCallRequest.RequestBasePrice,
		AdditionalPrice:             taxiCallTicket.DriverAdditionalPrice(),
		DriverAdditionalRewardPrice: taxiCallRequest.DriverAdditionalRewardPrice,
		Departure:                   taxiCallRequest.Departure,
		Arrival:                     taxiCallRequest.Arrival,
		ToArrivalRoute:              taxiCallRequest.ToArrivalRoute,
		ToDepartureDistance:         driverTaxiCallContext.ToDepartureDistance,
		Tags:                        taxiCallRequest.Tags,
		UserTag:                     taxiCallRequest.UserTag,
		UpdateTime:                  updateTime,
	}

	cmdJson, _ := json.Marshal(cmd)

	return entity.Event{
		MessageId:    utils.MustNewUUID(),
		EventUri:     EventUri_DriverTaxiCallNotification,
		DelaySeconds: 0,
		Payload:      cmdJson,
	}
}

func NewRawMessageCommand(accountId, category, messageTitle, messageBody string, data map[string]string) entity.Event {
	cmd := PushRawCommand{
		AccountId:    accountId,
		Category:     category,
		MessageTitle: messageTitle,
		MessageBody:  messageBody,
		Data:         data,
	}
	notificationJson, _ := json.Marshal(cmd)

	return entity.Event{
		MessageId:    utils.MustNewUUID(),
		EventUri:     EventUri_RawMessage,
		DelaySeconds: 0,
		Payload:      notificationJson,
	}
}
