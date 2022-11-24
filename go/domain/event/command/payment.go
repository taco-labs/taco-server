package command

import (
	"encoding/json"
	"fmt"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/utils"
)

var (
	EventUri_PaymentPrefix           = "Payment/"
	EventUri_UserDeletePayment       = fmt.Sprintf("%sUserRemovePayment", EventUri_PaymentPrefix)
	EventUri_UserTransaction         = fmt.Sprintf("%sUserTransaction", EventUri_PaymentPrefix)
	EventUri_UserTransactionFailed   = fmt.Sprintf("%sUserTransactionFailed", EventUri_PaymentPrefix)
	EventUri_UserTransactionRecovery = fmt.Sprintf("%sUserTransactionRecovery", EventUri_PaymentPrefix)
)

type PaymentUserPaymentDeleteCommand struct {
	UserId     string `json:"userId"`
	PaymentId  string `json:"paymentId"`
	BillingKey string `json:"billingKey"`
}

type PaymentUserTransactionCommand struct {
	UserId    string `json:"userId"`
	PaymentId string `json:"paymentId"`
	OrderId   string `json:"orderId"`
	OrderName string `json:"orderName"`
	Amount    int    `json:"amount"`
}

type PaymentUserTransactionFailedCommand struct {
	UserId             string `json:"userId"`
	PaymentId          string `json:"paymentId"`
	OrderId            string `json:"orderId"`
	OrderName          string `json:"orderName"`
	Amount             int    `json:"amount"`
	FailedErrorCode    string `json:"failedErrorCode"`
	FailedErrorMessage string `json:"failedErrorMessage"`
}

type PaymentUserTransactionRecoveryCommand struct {
	UserId    string `json:"userId"`
	PaymentId string `json:"paymentId"`
}

func NewPaymentUserPaymentDeleteCommand(userId string, paymentId string, billingKey string) entity.Event {
	cmd := PaymentUserPaymentDeleteCommand{
		UserId:     userId,
		PaymentId:  paymentId,
		BillingKey: billingKey,
	}

	cmdJson, _ := json.Marshal(cmd)

	return entity.Event{
		MessageId:    utils.MustNewUUID(),
		EventUri:     EventUri_UserDeletePayment,
		DelaySeconds: 0,
		Payload:      cmdJson,
	}
}

func NewPaymentUserTransactionCommand(userId string, paymentId string, orderId string, orderName string, amount int) entity.Event {
	cmd := PaymentUserTransactionCommand{
		UserId:    userId,
		PaymentId: paymentId,
		OrderId:   orderId,
		OrderName: orderName,
		Amount:    amount,
	}

	cmdJson, _ := json.Marshal(cmd)

	return entity.Event{
		MessageId:    utils.MustNewUUID(),
		EventUri:     EventUri_UserTransaction,
		DelaySeconds: 0,
		Payload:      cmdJson,
	}
}

func NewPaymentUserTransactionFailedCommand(
	userId string, paymentId string, orderId string, orderName string, amount int,
	failedErrorCode value.ErrCode, failedErrorMessage string,
) entity.Event {
	cmd := PaymentUserTransactionFailedCommand{
		UserId:             userId,
		PaymentId:          paymentId,
		OrderId:            orderId,
		OrderName:          orderName,
		Amount:             amount,
		FailedErrorCode:    string(failedErrorCode),
		FailedErrorMessage: failedErrorMessage,
	}

	cmdJson, _ := json.Marshal(cmd)

	return entity.Event{
		MessageId:    utils.MustNewUUID(),
		EventUri:     EventUri_UserTransactionFailed,
		DelaySeconds: 0,
		Payload:      cmdJson,
	}
}

func NewPaymentUserTransactionRecoveryCommand(userId string, paymentId string) entity.Event {
	cmd := PaymentUserTransactionRecoveryCommand{
		UserId:    userId,
		PaymentId: paymentId,
	}

	cmdJson, _ := json.Marshal(cmd)

	return entity.Event{
		MessageId:    utils.MustNewUUID(),
		EventUri:     EventUri_UserTransactionRecovery,
		DelaySeconds: 0,
		Payload:      cmdJson,
	}
}
