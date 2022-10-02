package request

import "github.com/taco-labs/taco/go/domain/value"

type CreateTaxiCallRequest struct {
	Dryrun    bool           `json:"dryrun"`
	Departure value.Location `json:"departure"`
	Arrival   value.Location `json:"arrival"`
	PaymentId string         `json:"paymentId"`
}
