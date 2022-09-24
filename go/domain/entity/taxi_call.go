package entity

import (
	"time"

	"github.com/taco-labs/taco/go/domain/value/enum"
	"github.com/uptrace/bun"
)

type TaxiCallRequest struct {
	bun.BaseModel `bun:"table:taxi_call_request"`

	Id                        string                    `bun:"id,pk"`
	UserId                    string                    `bun:"user_id"`
	DriverId                  string                    `bun:"driver_id"`
	DepartureLatitude         float32                   `bun:"departure_latitude"`
	DepartureLongitude        float32                   `bun:"departure_longitude"`
	ArrivalLatitude           float32                   `bun:"arrival_latitude"`
	ArrivalLongitude          float32                   `bun:"arrival_longitude"`
	PaymentId                 string                    `bun:"payment_id"`
	RequestBasePrice          int                       `bun:"request_base_price"`
	RequestMinAdditionalPrice int                       `bun:"request_min_additional_price"`
	RequestMaxAdditionalPrice int                       `bun:"request_max_additional_price"`
	BasePrice                 int                       `bun:"base_price"`
	AdditionalPrice           int                       `bun:"additional_price"`
	CallHistory               []*TaxiCallRequestHistory `bun:"rel:has-many,join:id=taxi_call_request_id"`
	CreateTime                time.Time                 `bun:"create_time"`
	UpdateTime                time.Time                 `bun:"update_time"`
}

type TaxiCallRequestHistory struct {
	bun.BaseModel `bun:"table:taxi_call_request_history"`

	Id                string             `bun:"id,pk"`
	TaxiCallRequestId string             `bun:"taxi_call_request_id"`
	TaxiCallState     enum.TaxiCallState `bun:"taxi_call_state"`
	CreateTime        time.Time          `bun:"create_time"`
}
