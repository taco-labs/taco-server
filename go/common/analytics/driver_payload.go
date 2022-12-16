package analytics

import (
	"time"

	"github.com/taco-labs/taco/go/domain/value"
)

type DriverSignupPayload struct {
	DriverId string
}

func (d DriverSignupPayload) EventType() EventType {
	return EventType_DriverSignup
}

type DriverLocationPayload struct {
	DriverId string
	Point    value.Point
}

func (d DriverLocationPayload) EventType() EventType {
	return EventType_DriverLocation
}

type DriverOnDutyPayload struct {
	DriverId string
	OnDuty   bool
}

func (d DriverOnDutyPayload) EventType() EventType {
	return EventType_DriverOnDuty
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
	DistanceToDeparture int
	DistanceToArrival   int
}

func (d DriverTaxicallTicketDistributionPayload) EventType() EventType {
	return EventType_DriverTaxiCallTicketDistribution
}

// TODO (taekyeom) user used point && additional reward
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

func (d DriverTaxiCallTicketAcceptPayload) EventType() EventType {
	return EventType_DriverTaxiCallTicketAccept
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

func (d DriverTaxiCallTicketRejectPayload) EventType() EventType {
	return EventType_DriverTaxiCallTicketReject
}

// TODO (taekyeom) Cancel penalty
type DriverTaxiCancelPayload struct {
	DriverId                  string
	RequestUserId             string
	TaxiCallRequestId         string
	DriverLocation            value.Point
	TaxiCallRequestCreateTime time.Time
	AcceptTime                time.Time
}

func (d DriverTaxiCancelPayload) EventType() EventType {
	return EventType_DriverTaxiCallCancel
}

type DriverTaxiToArrivalPayload struct {
	DriverId                  string
	RequestUserId             string
	TaxiCallRequestId         string
	DriverLocation            value.Point
	TaxiCallRequestCreateTime time.Time
	AcceptTime                time.Time
}

func (d DriverTaxiToArrivalPayload) EventType() EventType {
	return EventType_DriverTaxiToArrival
}

// TODO (taekyeom) toll fee & additional reward & referral point
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

func (d DriverTaxiDonePaylod) EventType() EventType {
	return EventType_DriverTaxiDone
}
