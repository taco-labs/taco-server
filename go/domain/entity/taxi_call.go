package entity

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/domain/value/enum"
	"github.com/taco-labs/taco/go/utils"
	"github.com/uptrace/bun"
)

const (
	AttemptLimit = 3
	PriceStep    = 1000
)

type TaxiCallRequest struct {
	bun.BaseModel `bun:"table:taxi_call_request"`

	// In Memroy
	Dryrun bool        `bun:"-"`
	Route  value.Route `bun:"-"`

	Id                        string               `bun:"id,pk"`
	UserId                    string               `bun:"user_id"`
	DriverId                  sql.NullString       `bun:"driver_id"`
	Departure                 value.Location       `bun:"departure,type:jsonb"`
	Arrival                   value.Location       `bun:"arrival,type:jsonb"`
	PaymentSummary            value.PaymentSummary `bun:"payment_summary,type:jsonb"`
	RequestBasePrice          int                  `bun:"request_base_price"`
	RequestMinAdditionalPrice int                  `bun:"request_min_additional_price"`
	RequestMaxAdditionalPrice int                  `bun:"request_max_additional_price"`
	BasePrice                 int                  `bun:"base_price"`
	AdditionalPrice           int                  `bun:"additional_price"`
	CurrentState              enum.TaxiCallState   `bun:"taxi_call_state"`
	CreateTime                time.Time            `bun:"create_time"`
	UpdateTime                time.Time            `bun:"update_time"`
}

// TODO (taekyeom) 취소 수수료 같은 로직을 나중에 고려해야 할듯
func (t *TaxiCallRequest) UpdateState(transitionTime time.Time, nextState enum.TaxiCallState) error {
	if !t.CurrentState.TryChangeState(nextState) {
		return value.ErrInvalidTaxiCallStateTransition
	}

	t.CurrentState = nextState
	t.UpdateTime = transitionTime

	return nil
}

type TaxiCallTicket struct {
	bun.BaseModel `bun:"table:taxi_call_ticket"`

	Id                string    `bun:"id,pk"`
	TaxiCallRequestId string    `bun:"taxi_call_request_id"`
	Attempt           int       `bun:"attempt"`
	AdditionalPrice   int       `bun:"additional_price"`
	CreateTime        time.Time `bun:"create_time"`
	UpdateTime        time.Time `bun:"update_time"`
}

func (t TaxiCallTicket) Copy() TaxiCallTicket {
	return TaxiCallTicket{
		Id:                t.Id,
		TaxiCallRequestId: t.TaxiCallRequestId,
		Attempt:           t.Attempt,
		AdditionalPrice:   t.AdditionalPrice,
		CreateTime:        t.CreateTime,
		UpdateTime:        t.UpdateTime,
	}
}

func (t TaxiCallTicket) ValidAttempt() bool {
	return t.Attempt <= AttemptLimit
}

func (t TaxiCallTicket) ValidAdditionalPrice(maxAdditionlPrice int) bool {
	return t.AdditionalPrice <= maxAdditionlPrice
}

func (t *TaxiCallTicket) IncreaseAttempt(updateTime time.Time) bool {
	t.Attempt += 1
	t.UpdateTime = updateTime
	return t.ValidAttempt()
}

func (t *TaxiCallTicket) IncreasePrice(maxPrice int, updateTime time.Time) bool {
	t.Id = utils.MustNewUUID()
	t.Attempt = 1
	t.AdditionalPrice += PriceStep
	t.CreateTime = updateTime
	t.UpdateTime = updateTime

	return t.ValidAdditionalPrice(maxPrice)
}

func (t TaxiCallTicket) GetRadius() int {
	switch t.Attempt {
	case 1:
		return 3000
	case 2:
		return 5000
	case 3:
		return 7000
	default:
		return 3000
	}
}

func (t TaxiCallTicket) GetRadiusMinutes() int {
	switch t.Attempt {
	case 1:
		return 3
	case 2:
		return 5
	case 3:
		return 7
	default:
		return 3
	}
}

type DriverTaxiCallContext struct {
	bun.BaseModel `bun:"table:driver_taxi_call_context"`

	DriverId                  string    `bun:"driver_id,pk"`
	CanReceive                bool      `bun:"can_receive"`
	LastReceivedRequestTicket string    `bun:"last_received_request_ticket"`
	RejectedLastRequestTicket bool      `bun:"rejected_last_request_ticket"`
	LastReceiveTime           time.Time `bun:"last_receive_time"`

	// Read Only
	Location value.Point `bun:"-"`
}

func NewEmptyDriverTaxiCallContext(driverId string, canReceive bool, t time.Time) DriverTaxiCallContext {
	return DriverTaxiCallContext{
		DriverId:                  driverId,
		CanReceive:                canReceive,
		LastReceivedRequestTicket: uuid.Nil.String(),
		RejectedLastRequestTicket: true,
		LastReceiveTime:           t,
	}
}

type DriverTaxiCallSettlement struct {
	bun.BaseModel `bun:"table:driver_taxi_call_settlement"`

	TaxiCallRequestId  string    `bun:"taxi_call_request_id,pk"`
	SettlementDone     bool      `bun:"settlement_done"`
	SettlementDoneTime time.Time `bun:"settlement_done_time"`
}
