package request

import (
	"net/url"
	"time"

	"github.com/taco-labs/taco/go/domain/value"
)

type ApplyDriverSettlementRequest struct {
	DriverId string
	OrderId  string
	Amount   int
}

type ListDriverSettlementHistoryRequest struct {
	DriverId  string `param:"driverId"`
	Count     int    `query:"count"`
	PageToken string `query:"pageToken"`
}

func (r ListDriverSettlementHistoryRequest) Validate() error {
	if r.PageToken != "" {
		timeStr, err := url.QueryUnescape(r.PageToken)
		if err != nil {
			return value.NewTacoError(value.ERR_INVALID, "Invalid url encoded param")
		}
		_, err = time.Parse(r.PageToken, timeStr)
		if err != nil {
			return value.NewTacoError(value.ERR_INVALID, "Invalid page token format")
		}
	}

	return nil
}

func (r ListDriverSettlementHistoryRequest) ToPageTokenTime() time.Time {
	if r.PageToken == "" {
		return time.Time{}
	}

	t, _ := time.Parse(r.PageToken, time.RFC3339Nano)
	return t
}

type DriverSettlementTransferSuccessCallbackRequest struct {
	DriverId      string
	Bank          string
	AccountNumber string
}

type DriverSettlementTransferFailureCallbackRequest struct {
	DriverId       string
	FailureMessage string
}

type ApplyDriverSettlementPromotionRewardRequest struct {
	DriverId   string
	OrderId    string
	Amount     int
	RewardRate int
}
