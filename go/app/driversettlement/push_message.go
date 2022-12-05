package driversettlement

import (
	"fmt"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/event/command"
	"github.com/taco-labs/taco/go/domain/value"
)

func settlementTransferFailureMessage(driverId string) entity.Event {
	return command.NewRawMessageCommand(
		driverId,
		value.NotificationCategory_Settlement,
		"타코 정산 요청이 실패하였습니다. 다시 시도해주세요.",
		"타코 정산 요청에 실패하였습니다. 정산 요청을 다시 시도하시거나, 지속적으로 실패하는 경우 고객센터에 문의해 주세요",
		map[string]string{},
	)
}

func settlementTransferSuccessMessage(driverId string, amount int) entity.Event {
	return command.NewRawMessageCommand(
		driverId,
		value.NotificationCategory_Settlement,
		fmt.Sprintf("타코 정산 요청 (%d 타코)가 완료되었습니다.", amount),
		fmt.Sprintf("정산 대상 타코 (%d 타코)가 정산 완료되어 기사님의 계좌로 지급되었습니다.", amount),
		map[string]string{},
	)
}
