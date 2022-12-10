package value

import (
	"time"

	"github.com/taco-labs/taco/go/domain/value/enum"
)

type DriverLatestTaxiCallTicket struct {
	TaxiCallRequestId string             `json:"taxiCallRequestId"`
	TaxiCallState     enum.TaxiCallState `json:"taxiCallState"`
	TaxiCallTicketId  string             `json:"taxiCallTicketId"`
	TicketAttempt     int                `json:"ticketAttempt"`
	RequestBasePrice  int                `json:"requestBasePrice"`
	AdditionalPrice   int                `json:"additionalPrice"`
	ToDeparture       Route              `json:"toDeparture"`
	ToArrival         Route              `json:"toArrival"`
	Tags              []string           `json:"tags"`
	UserTag           string             `json:"userTag"`
	UpdateTime        time.Time          `json:"updateTime"`
}
