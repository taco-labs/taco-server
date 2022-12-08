package command

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/utils"
)

var (
	EventUri_DriverSettlementPrefix            = "Settlement/"
	EventUri_DriverSettlementRequest           = fmt.Sprintf("%sRequest", EventUri_DriverSettlementPrefix)
	EventUri_DriverSettlementTransferRequest   = fmt.Sprintf("%sTransferRequest", EventUri_DriverSettlementPrefix)
	EventUri_DriverSettlementTransferExecution = fmt.Sprintf("%sTransferExecution", EventUri_DriverSettlementPrefix)
	EventUri_DriverSettlementTransferSuccess   = fmt.Sprintf("%sTransferSuccess", EventUri_DriverSettlementPrefix)
	EventUri_DriverSettlementTransferFail      = fmt.Sprintf("%sTransferFail", EventUri_DriverSettlementPrefix)
)

type DriverSettlementRequestCommand struct {
	DriverId          string    `json:"driverId"`
	TaxiCallRequestId string    `json:"taxiCallRequestId"`
	Amount            int       `json:"amount"`
	RequestTime       time.Time `json:"doneTime"`
}

type DriverSettlementTransferRequestCommand struct {
	DriverId string `json:"driverId"`
}

type DriverSettlementTransferExecutionCommand struct {
	DriverId string `json:"driverId"`
}

type DriverSettlementTransferSuccessCommand struct {
	DriverId      string `json:"driverId"`
	Bank          string `json:"bank"`
	AccountNumber string `json:"accountNumber"`
}

type DriverSettlementTransferFailCommand struct {
	DriverId       string `json:"driverId"`
	FailureMessage string `json:"failureMessage"`
}

func NewDriverSettlementRequestCommand(driverId, taxiCallRequestId string, amount int, requestTime time.Time) entity.Event {
	cmd := DriverSettlementRequestCommand{
		DriverId:          driverId,
		TaxiCallRequestId: taxiCallRequestId,
		Amount:            amount,
		RequestTime:       requestTime,
	}

	cmdJson, _ := json.Marshal(cmd)

	return entity.Event{
		MessageId:    utils.MustNewUUID(),
		EventUri:     EventUri_DriverSettlementRequest,
		DelaySeconds: 0,
		Payload:      cmdJson,
	}
}

func NewDriverSettlementTransferRequestCommand(driverId string) entity.Event {
	cmd := DriverSettlementTransferRequestCommand{driverId}

	cmdJson, _ := json.Marshal(cmd)

	return entity.Event{
		MessageId:    utils.MustNewUUID(),
		EventUri:     EventUri_DriverSettlementTransferRequest,
		DelaySeconds: 0,
		Payload:      cmdJson,
	}
}

func NewDriverSettlementTransferExecutionCommand(driverId string) entity.Event {
	cmd := DriverSettlementTransferExecutionCommand{driverId}

	cmdJson, _ := json.Marshal(cmd)

	return entity.Event{
		MessageId:    utils.MustNewUUID(),
		EventUri:     EventUri_DriverSettlementTransferExecution,
		DelaySeconds: 0,
		Payload:      cmdJson,
	}
}

func NewDriverSettlementTransferSuccessCommand(driverId, bank, accountNumber string) entity.Event {
	cmd := DriverSettlementTransferSuccessCommand{
		DriverId:      driverId,
		Bank:          bank,
		AccountNumber: accountNumber,
	}

	cmdJson, _ := json.Marshal(cmd)

	return entity.Event{
		MessageId:    utils.MustNewUUID(),
		EventUri:     EventUri_DriverSettlementTransferSuccess,
		DelaySeconds: 0,
		Payload:      cmdJson,
	}
}

func NewDriverSettlementTransferFailCommand(driverId string, failureMessage string) entity.Event {
	cmd := DriverSettlementTransferFailCommand{
		DriverId:       driverId,
		FailureMessage: failureMessage,
	}

	cmdJson, _ := json.Marshal(cmd)

	return entity.Event{
		MessageId:    utils.MustNewUUID(),
		EventUri:     EventUri_DriverSettlementTransferFail,
		DelaySeconds: 0,
		Payload:      cmdJson,
	}
}
