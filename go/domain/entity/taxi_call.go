package entity

import (
	"database/sql"
	"errors"
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

func CalculateDriverAdditionalPrice(additionalPrice int) int {
	return additionalPrice * 7 / 10
}

type DriverLatestTaxiCallRequest struct {
	TaxiCallRequest
	UserPhone string
}

type DriverLatestTaxiCallRequestTicket struct {
	DriverLatestTaxiCallRequest
	TicketId string
	Attempt  int
}

type UserLatestTaxiCallRequest struct {
	TaxiCallRequest
	DriverPhone     string
	DriverCarNumber string
}

type TaxiCallRequest struct {
	bun.BaseModel `bun:"table:taxi_call_request"`

	// In Memroy
	Dryrun bool     `bun:"-"`
	Tags   []string `bun:"-"`

	Id                        string               `bun:"id,pk"`
	UserId                    string               `bun:"user_id"`
	DriverId                  sql.NullString       `bun:"driver_id"`
	Departure                 value.Location       `bun:"departure,type:jsonb"`
	Arrival                   value.Location       `bun:"arrival,type:jsonb"`
	TagIds                    []int                `bun:"tag_ids,array"`
	UserTag                   string               `bun:"user_tag"`
	PaymentSummary            value.PaymentSummary `bun:"payment_summary,type:jsonb"`
	RequestBasePrice          int                  `bun:"request_base_price"`
	RequestMinAdditionalPrice int                  `bun:"request_min_additional_price"`
	RequestMaxAdditionalPrice int                  `bun:"request_max_additional_price"`
	BasePrice                 int                  `bun:"base_price"`
	TollFee                   int                  `bun:"toll_fee"`
	CancelPenaltyPrice        int                  `bun:"cancel_penalty_price"`
	AdditionalPrice           int                  `bun:"additional_price"`
	UserUsedPoint             int                  `bun:"user_used_point"`
	CurrentState              enum.TaxiCallState   `bun:"taxi_call_state"`
	ToDepartureRoute          value.Route          `bun:"to_departure_route"`
	ToArrivalRoute            value.Route          `bun:"to_arrival_route"`
	CreateTime                time.Time            `bun:"create_time"`
	UpdateTime                time.Time            `bun:"update_time"`
}

func (t TaxiCallRequest) TotalPrice() int {
	return t.BasePrice + t.TollFee + t.AdditionalPrice
}

func (t TaxiCallRequest) UserAdditionalPrice() int {
	return t.AdditionalPrice
}

func (t TaxiCallRequest) DriverSettlementAdditonalPrice() int {
	return CalculateDriverAdditionalPrice(t.AdditionalPrice)
}

func (t TaxiCallRequest) DriverSettlementCancelPenaltyPrice() int {
	return CalculateDriverAdditionalPrice(t.CancelPenaltyPrice)
}

// TODO (taekyeom) paramterize it
func (t TaxiCallRequest) UserCancelPenaltyPrice(cancelTime time.Time) int {
	cancelTimeHour := cancelTime.UTC().Hour()
	// 야간 (22시 ~ 새벽4시 in UTC)
	if cancelTimeHour >= 13 && cancelTimeHour < 19 {
		return 2000
	}
	return 1000
}

func (t TaxiCallRequest) DriverCancelPenaltyDuration() time.Duration {
	return time.Duration(0)
}

// TODO (taekyeom) 취소 수수료 같은 로직을 나중에 고려해야 할듯
func (t *TaxiCallRequest) UpdateState(transitionTime time.Time, nextState enum.TaxiCallState) error {
	if !t.CurrentState.TryChangeState(nextState) {
		return value.ErrInvalidTaxiCallStateTransition
	}

	if nextState == enum.TaxiCallState_USER_CANCELLED && t.userCancelNeedConfirmation(transitionTime) {
		return value.ErrConfirmationNeededStateTransition
	}

	if nextState == enum.TaxiCallState_DRIVER_CANCELLED && t.driverCancelNeedConfirmation() {
		return value.ErrConfirmationNeededStateTransition
	}

	t.CurrentState = nextState
	t.UpdateTime = transitionTime

	return nil
}

func (t *TaxiCallRequest) ForceUpdateState(transitionTime time.Time, nextState enum.TaxiCallState) error {
	err := t.UpdateState(transitionTime, nextState)
	if err != nil && !errors.Is(err, value.ErrConfirmationNeededStateTransition) {
		return err
	}

	t.CurrentState = nextState
	t.UpdateTime = transitionTime

	return nil
}

func (t TaxiCallRequest) userCancelNeedConfirmation(transitionTime time.Time) bool {
	return t.CurrentState == enum.TaxiCallState_DRIVER_TO_DEPARTURE && transitionTime.Sub(t.UpdateTime) > time.Minute
}

func (t TaxiCallRequest) driverCancelNeedConfirmation() bool {
	return t.CurrentState == enum.TaxiCallState_DRIVER_TO_DEPARTURE
}

type TaxiCallTicket struct {
	bun.BaseModel `bun:"table:taxi_call_ticket"`

	TaxiCallRequestId string    `bun:"taxi_call_request_id,pk"`
	AdditionalPrice   int       `bun:"additional_price,pk"`
	Attempt           int       `bun:"attempt,pk"`
	TicketId          string    `bun:"ticket_id"`
	DistributedCount  int       `bun:"distributed_count"`
	CreateTime        time.Time `bun:"create_time"`
}

func (t TaxiCallTicket) AttemptLimit() bool {
	return t.Attempt == AttemptLimit
}

func (t TaxiCallTicket) UserAdditionalPrice() int {
	return t.AdditionalPrice
}

func (t TaxiCallTicket) DriverAdditionalPrice() int {
	return CalculateDriverAdditionalPrice(t.AdditionalPrice)
}

func (t TaxiCallTicket) Step(maxPrice int, updateTime time.Time) (TaxiCallTicket, bool) {
	if t.Attempt < AttemptLimit {
		return TaxiCallTicket{
			TicketId:          t.TicketId,
			TaxiCallRequestId: t.TaxiCallRequestId,
			Attempt:           t.Attempt + 1,
			AdditionalPrice:   t.AdditionalPrice,
			CreateTime:        updateTime,
		}, true
	}

	if t.AdditionalPrice < maxPrice {
		additionalPrice := t.AdditionalPrice + PriceStep
		if additionalPrice > maxPrice {
			additionalPrice = maxPrice
		}
		return TaxiCallTicket{
			TicketId:          utils.MustNewUUID(),
			TaxiCallRequestId: t.TaxiCallRequestId,
			Attempt:           1,
			AdditionalPrice:   additionalPrice,
			CreateTime:        updateTime,
		}, true
	}

	return TaxiCallTicket{}, false
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
	BlockUntil                time.Time `bun:"block_until"`

	// Read Only
	Location            value.Point `bun:"-"`
	ToDepartureDistance int         `bun:"-"` // In meter
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

type DriverDenyTaxiCallTag struct {
	bun.BaseModel `bun:"table:driver_deny_taxi_call_tag"`

	DriverId string `bun:"driver_id,pk"`
	TagId    int    `bun:"tag_id,pk"`
	Tag      string `bun:"-"`
}
