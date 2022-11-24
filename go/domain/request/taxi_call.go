package request

import (
	"github.com/taco-labs/taco/go/domain/value"
)

type CreateTaxiCallRequest struct {
	Dryrun    bool        `json:"dryrun"`
	Departure value.Point `json:"departure"`
	Arrival   value.Point `json:"arrival"`
	PaymentId string      `json:"paymentId"`
	TagIds    []int       `json:"tagIds"`
}

// TODO (taekyeom) validation
func (c CreateTaxiCallRequest) Validate() error {
	return nil
}

type ListUserTaxiCallRequest struct {
	UserId    string `param:"userId"`
	Count     int    `query:"count"`
	PageToken string `query:"pageToken"`
}

type ListDriverTaxiCallRequest struct {
	DriverId  string `param:"driverId"`
	Count     int    `query:"count"`
	PageToken string `query:"pageToken"`
}

type DoneTaxiCallRequest struct {
	TaxiCallRequestId string `param:"taxiCallRequestId"`
	BasePrice         int    `json:"basePrice"`
}
