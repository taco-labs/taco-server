package request

import "time"

type ListDriverSettlementHistoryRequest struct {
	DriverId  string    `param:"driverId"`
	Count     int       `query:"count"`
	PageToken time.Time `query:"pageToken"`
}
