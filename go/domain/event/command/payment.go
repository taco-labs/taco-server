package command

import (
	"encoding/json"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/utils"
)

const (
	EventUri_UserTransaction = "Payment/UserTransaction"
)

type PaymentUserTransactionCommand struct {
	UserId    string `json:"userId"`
	PaymentId string `json:"paymentId"`
	OrderId   string `json:"orderId"`
	OrderName string `json:"orderName"`
	Amount    int    `json:"amount"`
}

func NewPaymentUserTransactionCommand(userId string, paymentId string, orderId string, orderName string, price int) entity.Event {
	cmd := PaymentUserTransactionCommand{
		UserId:    userId,
		PaymentId: paymentId,
		OrderId:   orderId,
		OrderName: orderName,
		Amount:    price,
	}

	cmdJson, _ := json.Marshal(cmd)

	return entity.Event{
		MessageId:    utils.MustNewUUID(),
		EventUri:     EventUri_UserTransaction,
		DelaySeconds: 0,
		Payload:      cmdJson,
	}
}
