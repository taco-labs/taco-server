package analytics

import (
	"time"

	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/domain/value/enum"
)

type UserSignupPayload struct {
	UserId       string
	ReferralType enum.ReferralType
	ReferralId   string
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
	UserId                    string
	Id                        string
	FailedTime                time.Time
	TaxiCallRequestCreateTime time.Time
}

func (u UserTaxiCallRequestFailedPayload) EventType() EventType {
	return EventType_UserTaxiCallRequestFailed
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
