package command

import (
	"encoding/json"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/utils"
)

const (
	EventUri_RawMessage                 = "Push/RawMessage"
	EventUri_UserTaxiCallNotification   = "Push/UserTaxiCall"
	EventUri_DriverTaxiCallNotification = "Push/DriverTaxiCall"
)

type PushUserTaxiCallCommand struct {
	UserId               string         `json:"userId"`
	TaxiCallRequestId    string         `json:"taxiCallRequestId"`
	TaxiCallState        string         `json:"taxiCallState"`
	RequestBasePrice     int            `json:"requestBasePrice,omitempty"`
	AdditionalPrice      int            `json:"additionalPrice,omitempty"`
	DriverId             string         `json:"driverId,omitempty"`
	BasePrice            int            `json:"basePrice,omitempty"`
	DriverLocation       value.Point    `json:"driverLocation,omitempty"`
	Departure            value.Location `json:"departureAddress,omitempty"`
	Arrival              value.Location `json:"arrivalAddress,omitempty"`
	SearchRangeInMinutes int            `json:"searchRangeInMinutes,omitempty"`
}

type PushDriverTaxiCallCommand struct {
	DriverId          string         `json:"driverId"`
	UserId            string         `json:"userId"`
	TaxiCallRequestId string         `json:"taxiCallRequestId"`
	TaxiCallTicketId  string         `json:"taxiCallTicketId"`
	TaxiCallState     string         `json:"taxiCallState"`
	DriverLocation    value.Point    `json:"driverLocation"`
	RequestBasePrice  int            `json:"requestBasePrice,omitempty"`
	AdditionalPrice   int            `json:"additionalPrice,omitempty"`
	Departure         value.Location `json:"departureAddress,omitempty"`
	Arrival           value.Location `json:"arrivalAddress,omitempty"`
	Attempt           int            `json:"attempt"`
}

func NewPushUserTaxiCallCommand(taxiCallRequest entity.TaxiCallRequest,
	taxiCallTicket entity.TaxiCallTicket,
	driverTaxiCallContext entity.DriverTaxiCallContext) entity.Event {
	cmd := PushUserTaxiCallCommand{
		UserId:               taxiCallRequest.UserId,
		TaxiCallRequestId:    taxiCallRequest.Id,
		TaxiCallState:        string(taxiCallRequest.CurrentState),
		RequestBasePrice:     taxiCallRequest.RequestBasePrice,
		AdditionalPrice:      taxiCallTicket.AdditionalPrice,
		DriverId:             taxiCallRequest.DriverId.String,
		BasePrice:            taxiCallRequest.BasePrice,
		DriverLocation:       driverTaxiCallContext.Location,
		Departure:            taxiCallRequest.Departure,
		Arrival:              taxiCallRequest.Arrival,
		SearchRangeInMinutes: taxiCallTicket.GetRadiusMinutes(),
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
	driverTaxiCallContext entity.DriverTaxiCallContext) entity.Event {
	cmd := PushDriverTaxiCallCommand{
		DriverId:          driverId,
		Attempt:           taxiCallTicket.Attempt,
		UserId:            taxiCallRequest.UserId,
		TaxiCallRequestId: taxiCallRequest.Id,
		TaxiCallTicketId:  taxiCallTicket.TicketId,
		TaxiCallState:     string(taxiCallRequest.CurrentState),
		DriverLocation:    driverTaxiCallContext.Location,
		RequestBasePrice:  taxiCallRequest.RequestBasePrice,
		AdditionalPrice:   taxiCallTicket.AdditionalPrice,
		Departure:         taxiCallRequest.Departure,
		Arrival:           taxiCallRequest.Arrival,
	}

	cmdJson, _ := json.Marshal(cmd)

	return entity.Event{
		MessageId:    utils.MustNewUUID(),
		EventUri:     EventUri_DriverTaxiCallNotification,
		DelaySeconds: 0,
		Payload:      cmdJson,
	}
}

func NewRawMessageCommand(notification value.Notification) entity.Event {
	notificationJson, _ := json.Marshal(notification)

	return entity.Event{
		MessageId:    utils.MustNewUUID(),
		EventUri:     EventUri_RawMessage,
		DelaySeconds: 0,
		Payload:      notificationJson,
	}
}
