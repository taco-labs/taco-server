package command

import (
	"encoding/json"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/utils"
)

const (
	EventUri_UserTaxiCallNotification   = "TaxiCallNotification/User"
	EventUri_DriverTaxiCallNotification = "TaxiCallNotification/Driver"
)

type UserTaxiCallNotificationCommand struct {
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

type DriverTaxiCallNotificationCommand struct {
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
}

func NewUserTaxiCallNotificationCommand(taxiCallRequest entity.TaxiCallRequest,
	taxiCallTicket entity.TaxiCallTicket,
	driverTaxiCallContext entity.DriverTaxiCallContext) entity.Event {
	cmd := UserTaxiCallNotificationCommand{
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

func NewDriverTaxiCallNotificationCommand(taxiCallRequest entity.TaxiCallRequest,
	taxiCallTicket entity.TaxiCallTicket,
	driverTaxiCallContext entity.DriverTaxiCallContext) entity.Event {
	cmd := DriverTaxiCallNotificationCommand{
		DriverId:          driverTaxiCallContext.DriverId,
		UserId:            taxiCallRequest.UserId,
		TaxiCallRequestId: taxiCallRequest.Id,
		TaxiCallTicketId:  taxiCallTicket.Id,
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
