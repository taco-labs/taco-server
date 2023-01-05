package payment

import (
	"fmt"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/event/command"
	"github.com/taco-labs/taco/go/domain/value"
)

func NewPaymentFallbackNotification(
	failedUserPayment entity.UserPayment, fallbackUserPayment entity.UserPayment,
) entity.Event {
	messageTitle := "타코 서비스에 대한 이용 요금 결제에 실패했습니다."
	messageBody := fmt.Sprintf("결제 수단 %s 대신 등록한 다른 결제 수단 (%s) 으로 결제를 시도합니다.",
		failedUserPayment.Name, fallbackUserPayment.Name)

	return command.NewRawMessageCommand(failedUserPayment.UserId, value.NotificationCategory_Payment, messageTitle, messageBody, map[string]string{})
}

func NewPaymentFailedNotification(userId string) entity.Event {
	messageTitle := "타코 서비스에 대한 이용 요금 결제에 실패했습니다. 결제 수단을 확인해주세요"
	messageBody := "고객님께서 등록하신 모든 결제 수단에 대한 이용 요금 결제에 실패했습니다. 새로운 결제 수단을 등록하거나 기존 결제 수단을 확인 후 복구를 진행해주세요."

	return command.NewRawMessageCommand(userId, value.NotificationCategory_Payment, messageTitle, messageBody, map[string]string{})
}

func NewUserReferralRewardNotification(userId string, amount int) entity.Event {
	messageTitle := fmt.Sprintf("추천인 택시 탑승 적립금 %d타코가 적립되었습니다.", amount)
	messageBody := ""

	return command.NewRawMessageCommand(userId, value.NotificationCategory_Payment, messageTitle, messageBody, map[string]string{})
}

func NewDriverReferralRewardNotification(driverId string, amount int) entity.Event {
	messageTitle := fmt.Sprintf("추천인 택시 운행 적립금 %d타코가 적립되었습니다.", amount)
	messageBody := ""

	return command.NewRawMessageCommand(driverId, value.NotificationCategory_Payment, messageTitle, messageBody, map[string]string{})
}
