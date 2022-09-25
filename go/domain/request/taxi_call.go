package request

import "github.com/taco-labs/taco/go/domain/value"

type CreateTaxiCallRequest struct {
	Departure        value.Location `json:"departure"`
	Arrival          value.Location `json:"arrival"`
	PaymentId        string         `json:"paymentId"`
	RequestBasePrice int            `json:"requestBasePrice"`
}
