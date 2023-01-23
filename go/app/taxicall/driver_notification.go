package taxicall

import (
	"time"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/event/command"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/domain/value/enum"
)

func newTaxiCallRequestInvalidateCommand(driverId, taxiCallRequestId string, updateTime time.Time) entity.Event {
	data := map[string]string{
		"taxiCallRequestId": taxiCallRequestId,
		"taxiCallState":     string(enum.TaxiCallState_REQUEST_INVALIDATED),
		"updateTime":        updateTime.Format(time.RFC3339),
	}

	return command.NewRawMessageCommand(driverId, value.NotificationCategory_Taxicall, "", "", data)
}
