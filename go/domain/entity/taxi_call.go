package entity

import (
	"time"

	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/domain/value/enum"
	"github.com/uptrace/bun"
)

type TaxiCallRequest struct {
	bun.BaseModel `bun:"table:taxi_call_request"`

	Id                        string                         `bun:"id,pk"`
	UserId                    string                         `bun:"user_id"`
	DriverId                  string                         `bun:"driver_id"`
	Departure                 value.Location                 `bun:"departure,type:jsonb"`
	Arrival                   value.Location                 `bun:"arrival,type:jsonb"`
	PaymentSummary            value.PaymentSummary           `bun:"payment_summary,type:jsonb"`
	RequestBasePrice          int                            `bun:"request_base_price"`
	RequestMinAdditionalPrice int                            `bun:"request_min_additional_price"`
	RequestMaxAdditionalPrice int                            `bun:"request_max_additional_price"`
	BasePrice                 int                            `bun:"base_price"`
	AdditionalPrice           int                            `bun:"additional_price"`
	CurrentState              enum.TaxiCallState             `bun:"taxi_call_state,type:jsonb"`
	CallHistory               []value.TaxiCallRequestHistory `bun:"taxi_call_state_history"`
	CreateTime                time.Time                      `bun:"create_time"`
	UpdateTime                time.Time                      `bun:"update_time"`
}

// TODO (update state)
