package payment

import (
	"fmt"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/event/command"
	"github.com/taco-labs/taco/go/domain/value"
)

func NewPaymentSuccessNotification(paymentOrder entity.UserPaymentOrder) entity.Event {
	messageTitle := fmt.Sprintf("타코 서비스에 대한 이요 요금 %d 원이 결제 결제되었습니다.", paymentOrder.Amount)

	return command.NewRawMessageCommand(paymentOrder.UserId, value.NotificationCategory_Payment, messageTitle, "", map[string]string{})
}

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
