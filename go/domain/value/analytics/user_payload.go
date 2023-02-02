package analytics

import (
	"time"

	"github.com/taco-labs/taco/go/domain/value"
)

type UserSignupPayload struct {
	UserId     string
	ReferralId string
}

func (u UserSignupPayload) EventType() EventType {
	return EventType_UserSignup
}

type UserTaxiCallRequestPayload struct {
	UserId                    string
	Id                        string
	Departure                 value.Location
	Arrival                   value.Location
	ToArrivalETA              time.Duration
	ToArrivalDistance         int
	TagIds                    []int
	Tags                      []string
	UserTag                   string
	PaymentSummary            value.PaymentSummary
	RequestBasePrice          int
	RequestMinAdditionalPrice int
	RequestMaxAdditionalPrice int
}

func (u UserTaxiCallRequestPayload) EventType() EventType {
	return EventType_UserTaxiCallRequest
}

type UserTaxiCallRequestFailedPayload struct {
	Id         string
	UserId     string
	FailedTime time.Time
	// TODO (taekyeom) last ticket context
	TaxiCallRequestCreateTime time.Time
}

func (u UserTaxiCallRequestFailedPayload) EventType() EventType {
	return EventType_UserTaxiCallRequestFailed
}

// TODO (taekyeom) ticket attempt analytics

type UserTaxiCallRequestDriverNotAvailablePayload struct {
	Id                        string
	UserId                    string
	FailedTime                time.Time
	TaxiCallRequestCreateTime time.Time
}

func (u UserTaxiCallRequestDriverNotAvailablePayload) EventType() EventType {
	return EventType_UserTaxiCallRequestDriverNotAvailable
}

type UserCancelTaxiCallRequestPayload struct {
	UserId                    string
	Id                        string
	CancelPenalty             int
	TaxiCallRequestCreateTime time.Time
}

func (u UserCancelTaxiCallRequestPayload) EventType() EventType {
	return EventType_UserCancelTaxiCallRequest
}

type UserReferralPointReceivedPayload struct {
	UserId        string
	FromUserId    string
	OrderId       string
	ReceiveAmount int
}

func (u UserReferralPointReceivedPayload) EventType() EventType {
	return EventType_UserReferralPointReceived
}

type UserAccessPayload struct {
	UserId string
	Method string
	Path   string
}

func (u UserAccessPayload) EventType() EventType {
	return EventType_UserAccess
}
