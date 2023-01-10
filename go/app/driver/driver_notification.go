package driver

import (
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/event/command"
	"github.com/taco-labs/taco/go/domain/value"
)

func NewDriverActivatedNotification(driverId string) entity.Event {
	messageTitle := "타코택시 기사님 가입이 승인되었습니다."
	messageBody := "타코택시 기사용 앱을 사용하실 수 있습니다. 출근 버튼을 눌러 콜을 수신하세요."

	return command.NewRawMessageCommand(driverId, value.NotificationCategory_Driver, messageTitle, messageBody, map[string]string{})
}
