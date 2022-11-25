package entity

import (
	"time"

	"github.com/uptrace/bun"
)

type DriverSettlementRequest struct {
	bun.BaseModel `bun:"table:driver_settlement_request"`

	TaxiCallRequestId string    `bun:"taxi_call_request_id,pk"`
	DriverId          string    `bun:"driver_id"`
	Amount            int       `bun:"amount"`
	CreateTime        time.Time `bun:"create_time"` // TODO (taekyeom) 타이밍 이슈가 없으려면 해당 시각이 결제 완료 시각과 일치해야 함
}

type DriverExpectedSettlement struct {
	bun.BaseModel `bun:"table:driver_expected_settlement"`

	DriverId       string `bun:"driver_id,pk"`
	ExpectedAmount int    `bun:"expected_amount"`
}

type DriverSettlementHistory struct {
	bun.BaseModel `bun:"table:driver_settlement_history"`

	DriverId              string    `bun:"driver_id,pk"`
	SettlementPeriodStart time.Time `bun:"settlement_period_start"`
	SettlementPeriodEnd   time.Time `bun:"settlement_period_end"`
	CreateTime            time.Time `bun:"create_time"`
	Amount                int       `bun:"amount"`
}
