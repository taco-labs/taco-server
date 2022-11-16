package command

import (
	"encoding/json"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/utils"
)

const (
	EventUri_UserTransaction       = "Payment/UserTransaction"
	EventUri_UserTransactionFailed = "Payment/UserTransactionFailed"
)

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
