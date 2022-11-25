package command

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/utils"
)

var (
	EventUri_DriverSettlementPrefix  = "Settlement/"
	EventUri_DriverSettlementRequest = fmt.Sprintf("%sRequest", EventUri_DriverSettlementPrefix)
	EventUri_DriverSettlementDone    = fmt.Sprintf("%sDone", EventUri_DriverSettlementPrefix)
)

type DriverSettlementRequestCommand struct {
	DriverId          string    `json:"driverId"`
	TaxiCallRequestId string    `json:"taxiCallRequestId"`
	Amount            int       `json:"amount"`
	RequestTime       time.Time `json:"doneTime"`
}

type DriverSettlementDoneCommand struct {
	DriverId              string    `json:"driverId"`
	Amount                int       `json:"amount"`
	SettlementPeriodStart time.Time `json:"settlement_period_start"`
	SettlementPeriodEnd   time.Time `json:"settlement_period_end"`
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

func NewDriverSettlementDoneCommand(driverId string, periodStart, periodEnd time.Time, amount int) entity.Event {
	cmd := DriverSettlementDoneCommand{
		DriverId:              driverId,
		Amount:                amount,
		SettlementPeriodStart: periodStart,
		SettlementPeriodEnd:   periodEnd,
	}

	cmdJson, _ := json.Marshal(cmd)

	return entity.Event{
		MessageId:    utils.MustNewUUID(),
		EventUri:     EventUri_DriverSettlementDone,
		DelaySeconds: 0,
		Payload:      cmdJson,
	}
}
