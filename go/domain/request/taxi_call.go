package request

import (
	"github.com/taco-labs/taco/go/domain/value"
)

type CreateTaxiCallRequest struct {
	Dryrun    bool        `json:"dryrun"`
	Departure value.Point `json:"departure"`
	Arrival   value.Point `json:"arrival"`
	PaymentId string      `json:"paymentId"`
}

// TODO (taekyeom) validation
func (c CreateTaxiCallRequest) Validate() error {
	return nil
}
