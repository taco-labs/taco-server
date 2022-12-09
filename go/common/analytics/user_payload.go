package analytics

import (
	"time"

	"github.com/taco-labs/taco/go/domain/value"
)

type UserSignupPayload struct {
	UserId string
}

type UserTaxiCallRequestPayload struct {
	UserId                    string
	Id                        string
	Departure                 value.Location
	Arrival                   value.Location
	ToArrivalETA              time.Duration
	ToArrivalDistance         int
	Tags                      []string
	UserTag                   string
	PaymentSummary            value.PaymentSummary
	RequestBasePrice          int
	RequestMinAdditionalPrice int
	RequestMaxAdditionalPrice int
}

type UserTaxiCallRequestFailedPayload struct {
	UserId                    string
	Id                        string
	FailedTime                time.Time
	TaxiCallRequestCreateTime time.Time
}

type UserCancelTaxiCallRequestPayload struct {
	UserId        string
	Id            string
	CancelPanelty int
	CreateTime    time.Time
}

type UserPaymentDonePayload struct {
	UserId         string
	OrderId        string
	OrderName      string
	Amount         int
	PaymentSummary value.PaymentSummary
}
