package command

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/utils"
)

var (
	EventUri_PaymentPrefix = "Payment/"

	EventUri_UserTransactionRequest = fmt.Sprintf("%sUserTransactionRequest", EventUri_PaymentPrefix)
	EventUri_UserTransactionSuccess = fmt.Sprintf("%sUserTransactionSuccess", EventUri_PaymentPrefix)
	EventUri_UserTransactionFail    = fmt.Sprintf("%sUserTransactionFail", EventUri_PaymentPrefix)

	EventUri_UserDeletePayment = fmt.Sprintf("%sUserRemovePayment", EventUri_PaymentPrefix)
)

type PaymentUserTransactionRequestCommand struct {
	UserId             string `json:"userId"`
	PaymentId          string `json:"paymentId"`
	OrderId            string `json:"orderId"`
	OrderName          string `json:"orderName"`
	Amount             int    `json:"amount"`
	UsedPoint          int    `json:"usedPoint"`
	SettlementTargetId string `json:"settlementTargetId"`
	SettlementAmount   int    `json:"settlementAmount"`
	Recovery           bool   `json:"recovery"`
}

func (p PaymentUserTransactionRequestCommand) TransactionAmount() int {
	return p.Amount - p.UsedPoint
}

type PaymentUserTransactionSuccessCommand struct {
	OrderId    string    `json:"orderId"`
	PaymentKey string    `json:"paymentKey"`
	ReceiptUrl string    `json:"receiptUrl"`
	CreateTime time.Time `json:"createTime"`
}

type PaymentUserTransactionFailCommand struct {
	OrderId       string `json:"orderId"`
	FailureCode   string `json:"failureCode"`
	FailureReason string `json:"failureReason"`
}

type PaymentUserPaymentDeleteCommand struct {
	UserId     string `json:"userId"`
	PaymentId  string `json:"paymentId"`
	BillingKey string `json:"billingKey"`
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

func NewUserPaymentTransactionRequestCommand(userId, paymentId, orderId, orderName, settlementTargetId string, amount,
	usedPoint, settlementAmount int, inRecovery bool) entity.Event {
	cmd := PaymentUserTransactionRequestCommand{
		UserId:             userId,
		PaymentId:          paymentId,
		OrderId:            orderId,
		OrderName:          orderName,
		Amount:             amount,
		UsedPoint:          usedPoint,
		SettlementTargetId: settlementTargetId,
		SettlementAmount:   settlementAmount,
		Recovery:           inRecovery,
	}

	cmdJson, _ := json.Marshal(cmd)

	return entity.Event{
		MessageId:    utils.MustNewUUID(),
		EventUri:     EventUri_UserTransactionRequest,
		DelaySeconds: 0,
		Payload:      cmdJson,
	}
}

func NewUserPaymentTransactionSuccessCommand(orderId, paymentKey, receiptUrl string, createTime time.Time) entity.Event {
	cmd := PaymentUserTransactionSuccessCommand{
		OrderId:    orderId,
		PaymentKey: paymentKey,
		ReceiptUrl: receiptUrl,
		CreateTime: createTime,
	}

	cmdJson, _ := json.Marshal(cmd)

	return entity.Event{
		MessageId:    utils.MustNewUUID(),
		EventUri:     EventUri_UserTransactionSuccess,
		DelaySeconds: 0,
		Payload:      cmdJson,
	}
}

func NewUserPaymentTransactionFailCommand(orderId, failureCode, failureReason string) entity.Event {
	cmd := PaymentUserTransactionFailCommand{
		OrderId:       orderId,
		FailureCode:   failureCode,
		FailureReason: failureReason,
	}

	cmdJson, _ := json.Marshal(cmd)

	return entity.Event{
		MessageId:    utils.MustNewUUID(),
		EventUri:     EventUri_UserTransactionFail,
		DelaySeconds: 0,
		Payload:      cmdJson,
	}
}
