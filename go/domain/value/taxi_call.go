package value

import (
	"time"

	"github.com/taco-labs/taco/go/domain/value/enum"
)

type TaxiCallRequestHistory struct {
	TaxiCallState enum.TaxiCallState `json:"taxiCallState"`
	CreateTime    time.Time          `json:"createTime"`
}
