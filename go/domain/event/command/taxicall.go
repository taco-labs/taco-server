package command

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value/enum"
	"github.com/taco-labs/taco/go/utils"
)

var (
	EventUri_TaxiCallPrefix            = "TaxiCall/"
	EventUri_TaxiCallProcess           = fmt.Sprintf("%sProcess", EventUri_TaxiCallPrefix)
	EventUri_TaxiCallInDirivngLocation = fmt.Sprintf("%sInDrivingLocation", EventUri_TaxiCallPrefix)
)

type TaxiCallProcessMessage struct {
	TaxiCallRequestId   string    `json:"taxiCallRequestId"`
	TaxiCallState       string    `json:"taxiCallState"`
	EventTime           time.Time `json:"eventTime"`
	DesiredScheduleTime time.Time `json:"desiredScheduleTime"`
}

func (t TaxiCallProcessMessage) ToEvent() entity.Event {
	payload, _ := json.Marshal(t)
	var delaySeconds int32
	if t.DesiredScheduleTime == t.EventTime {
		delaySeconds = 0
	} else {
		delayDuration := minTimeDuration(
			t.DesiredScheduleTime.Sub(t.EventTime),
			time.Until(t.DesiredScheduleTime),
		)
		if delayDuration < 0 {
			delayDuration = 0
		}
		delaySeconds = int32(delayDuration.Seconds())
	}
	return entity.Event{
		MessageId:    utils.MustNewUUID(),
		EventUri:     EventUri_TaxiCallProcess,
		DelaySeconds: delaySeconds,
		Payload:      payload,
		CreateTime:   t.EventTime,
	}
}

type TaxiCallInDrivingLocationMessage struct {
	TaxiCallRequestId   string    `json:"taxiCallRequestId"`
	DesiredScheduleTime time.Time `json:"desiredScheduleTime"`
	EventTime           time.Time `json:"eventTIme"`
}

func (t TaxiCallInDrivingLocationMessage) ToEvent() entity.Event {
	payload, _ := json.Marshal(t)
	var delaySeconds int32
	if t.DesiredScheduleTime == t.EventTime {
		delaySeconds = 0
	} else {
		delayDuration := minTimeDuration(
			t.DesiredScheduleTime.Sub(t.EventTime),
			time.Until(t.DesiredScheduleTime),
		)
		if delayDuration < 0 {
			delayDuration = 0
		}
		delaySeconds = int32(delayDuration.Seconds())
	}
	return entity.Event{
		MessageId:    utils.MustNewUUID(),
		EventUri:     EventUri_TaxiCallInDirivngLocation,
		DelaySeconds: delaySeconds,
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

func NewTaxiCallInDrivingLocationMessage(taxiCallRequestId string,
	eventTime time.Time, desiredScheduleTime time.Time) entity.Event {
	command := TaxiCallInDrivingLocationMessage{
		TaxiCallRequestId:   taxiCallRequestId,
		EventTime:           eventTime,
		DesiredScheduleTime: desiredScheduleTime.Truncate(time.Microsecond),
	}

	return command.ToEvent()
}

func minTimeDuration(t1 time.Duration, t2 time.Duration) time.Duration {
	if t1 < t2 {
		return t1
	}
	return t2
}
