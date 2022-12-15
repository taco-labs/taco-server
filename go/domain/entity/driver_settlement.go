package entity

import (
	"fmt"
	"time"

	"github.com/taco-labs/taco/go/domain/value/enum"
	"github.com/uptrace/bun"
)

const (
	DriverSettlementTaxRateNumerator   = 33
	DriverSettlementTaxRateDenominator = 1000
)

func ExpectedSettlementAmountWithoutTax(amount int) int {
	return amount * (DriverSettlementTaxRateDenominator - DriverSettlementTaxRateNumerator) / DriverSettlementTaxRateDenominator
}

type DriverSettlementRequest struct {
	bun.BaseModel `bun:"table:driver_settlement_request"`

	TaxiCallRequestId string    `bun:"taxi_call_request_id,pk"`
	DriverId          string    `bun:"driver_id"`
	Amount            int       `bun:"amount"`
	CreateTime        time.Time `bun:"create_time"` // TODO (taekyeom) 타이밍 이슈가 없으려면 해당 시각이 결제 완료 시각과 일치해야 함
}

type DriverTotalSettlement struct {
	bun.BaseModel `bun:"table:driver_total_settlement"`

	DriverId          string `bun:"driver_id,pk"`
	TotalAmount       int    `bun:"total_amount"`
	RequestableAmount int    `bun:"-"`
}

type DriverSettlementHistory struct {
	bun.BaseModel `bun:"table:driver_settlement_history"`

	DriverId         string    `bun:"driver_id,pk"`
	Amount           int       `bun:"amount"`
	AmountWithoutTax int       `bun:"amoun_without_tax"`
	Bank             string    `bun:"bank"`
	AccountNumber    string    `bun:"account_number"`
	RequestTime      time.Time `bun:"request_time"`
	CreateTime       time.Time `bun:"create_time"`
}

func (d DriverSettlementHistory) RedactedAccountNumber() string {
	lastAccountNumber := d.AccountNumber[len(d.AccountNumber)-4:]
	return fmt.Sprintf("****%s", lastAccountNumber)
}

type DriverInflightSettlementTransfer struct {
	bun.BaseModel `bun:"table:driver_inflight_settlement_transfer"`

	TransferId        string                              `bun:"transfer_id,pk"`
	DriverId          string                              `bun:"driver_id"`
	ExecutionKey      string                              `bun:"execution_key"`
	BankTransactionId string                              `bun:"bank_transaction_id"`
	Amount            int                                 `bun:"amount"`
	AmountWithoutTax  int                                 `bun:"amoun_without_tax"`
	Message           string                              `bun:"message"`
	State             enum.SettlementTransferProcessState `bun:"state"`
	CreateTime        time.Time                           `bun:"create_time"`
	UpdateTime        time.Time                           `bun:"update_time"`
}

type DriverFailedSettlementTransfer struct {
	bun.BaseModel `bun:"table:driver_failed_settlement_transfer"`

	TransferId        string    `bun:"transfer_id,pk"`
	DriverId          string    `bun:"driver_id"`
	ExecutionKey      string    `bun:"execution_key"`
	BankTransactionId string    `bun:"bank_transaction_id"`
	Amount            int       `bun:"amount"`
	AmountWithoutTax  int       `bun:"amoun_without_tax"`
	Message           string    `bun:"message"`
	FailureMessage    string    `bun:"failure_message"`
	CreateTime        time.Time `bun:"create_time"`
}
