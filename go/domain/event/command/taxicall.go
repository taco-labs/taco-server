package command

import (
	"encoding/json"
	"math"
	"time"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value/enum"
	"github.com/taco-labs/taco/go/utils"
)

const (
	EventUri_TaxiCallProcess = "TaxiCall/Process"
)

type TaxiCallProcessMessage struct {
	TaxiCallRequestId   string    `json:"taxiCallRequestId"`
	TaxiCallState       string    `json:"taxiCallState"`
	EventTime           time.Time `json:"eventTime"`
	DesiredScheduleTime time.Time `json:"desiredScheduleTime"`
}

func (t TaxiCallProcessMessage) ToEvent() entity.Event {
	payload, _ := json.Marshal(t)
	var delaySeconds float64
	if t.DesiredScheduleTime == t.EventTime {
		delaySeconds = 0
	} else {
		delayMin := math.Min(
			float64(t.DesiredScheduleTime.Sub(t.EventTime)),
			float64(time.Until(t.DesiredScheduleTime)),
		)
		delaySeconds = math.Min(delayMin, 0)
	}
	return entity.Event{
		MessageId:    utils.MustNewUUID(),
		EventUri:     EventUri_TaxiCallProcess,
		DelaySeconds: int64(delaySeconds),
		Payload:      payload,
		CreateTime:   t.EventTime,
	}
}

func NewTaxiCallProgressCommand(taxiCallRequestId string, taxiCallState enum.TaxiCallState,
	eventTime time.Time, desiredProcessTime time.Time) entity.Event {
	command := TaxiCallProcessMessage{
		TaxiCallRequestId:   taxiCallRequestId,
		TaxiCallState:       string(taxiCallState),
		EventTime:           eventTime.Truncate(time.Microsecond),          // (taekyeom) postgresql이 microsecond까지만 지원
		DesiredScheduleTime: desiredProcessTime.Truncate(time.Microsecond), // (taekyeom) postgresql이 microsecond까지만 지원
	}

	return command.ToEvent()
}
