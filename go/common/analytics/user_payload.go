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
	ETA                       time.Duration
	Distance                  int
	Tags                      []string
	UserTag                   string
	PaymentSummary            value.PaymentSummary
	RequestBasePrice          int
	RequestMinAdditionalPrice int
	RequestMaxAdditionalPrice int
}

type UserCancelTaxiCallRequestPayload struct {
	UserId     string
	Id         string
	CreateTime time.Time
}

type UserPaymentDonePayload struct {
	UserId         string
	OrderId        string
	OrderName      string
	Amount         int
	PaymentSummary value.PaymentSummary
}
