package analytics

import (
	"time"

	"github.com/taco-labs/taco/go/domain/value"
)

type DriverSignupPayload struct {
	DriverId string
}

type DriverLocationPayload struct {
	DriverId string
	Point    value.Point
}

type DriverOnDutyPayload struct {
	DriverId string
	OnDuty   bool
}

// TODO (taekyeom) ETA to departure
type DriverTaxicallTicketDistributionPayload struct {
	DriverId                  string
	RequestUserId             string
	TaxiCallRequestId         string
	TaxiCallRequestTicketId   string
	TicketAttempt             int
	RequestBasePrice          int
	AdditionalPrice           int
	DriverLocation            value.Point
	TaxiCallRequestCreateTime time.Time
	// ETAToArrival              time.Duration  // TODO (taekyeom) add eta to arrival & distance
	// DistanceToArrival         int
}

type DriverTaxiCallTicketAcceptPayload struct {
	DriverId                        string
	RequestUserId                   string
	TaxiCallRequestId               string
	ReceivedTaxiCallRequestTicketId string
	ReceivedTicketAttempt           int
	ActualTaxiCallRequestTicketId   string
	ActualTicketAttempt             int
	RequestBasePrice                int
	AdditionalPrice                 int
	DriverLocation                  value.Point
	ReceiveTime                     time.Time
	TaxiCallRequestCreateTime       time.Time
}

type DriverTaxiCallTicketRejectPayload struct {
	DriverId                  string
	RequestUserId             string
	TaxiCallRequestId         string
	TaxiCallRequestTicketId   string
	TicketAttempt             int
	RequestBasePrice          int
	AdditionalPrice           int
	DriverLocation            value.Point
	ReceiveTime               time.Time
	TaxiCallRequestCreateTime time.Time
}

// TODO (taekyeom) Cancel
type DriverTaxiCancelPayload struct {
	DriverId                  string
	RequestUserId             string
	TaxiCallRequestId         string
	DriverLocation            value.Point
	TaxiCallRequestCreateTime time.Time
	AcceptTime                time.Time
}

type DriverTaxiToArrivalPayload struct {
	DriverId                  string
	RequestUserId             string
	TaxiCallRequestId         string
	DriverLocation            value.Point
	TaxiCallRequestCreateTime time.Time
	AcceptTime                time.Time
}

type DriverTaxiDonePaylod struct {
	DriverId                  string
	RequestUserId             string
	TaxiCallRequestId         string
	BasePrice                 int
	RequestBasePrice          int
	AdditionalPrice           int
	DriverLocation            value.Point
	TaxiCallRequestCreateTime time.Time
	ToArrivalTime             time.Time
}
