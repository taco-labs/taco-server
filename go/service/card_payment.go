package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/value"
)

const (
	tossPaymentCardReigstrationPath  = "v1/billing/authorizations/card"
	tossPaymentTransactionPath       = "v1/billing/%s"
	tossPaymentCancelTransactionPath = "v1/payments/%s/cancel"
	tossPaymentGetTransaction        = "v1/payments/orders/%s"
)

type PaymentService interface {
	RegisterCard(context.Context, string, request.UserPaymentRegisterRequest) (value.CardPaymentInfo, error)
	Transaction(context.Context, entity.UserPayment, value.Payment) (value.PaymentResult, error) // TODO(taekyeom) 결제 기록 별도 보관 필요
	CancelTransaction(context.Context, value.PaymentCancel) error
	GetTransactionResult(context.Context, string) (value.PaymentResult, error)
}

type tossPaymentService struct {
	client *resty.Client
}

type tossPaymentServiceError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type tossPaymentCardRegisterRequest struct {
	CustomerKey            string `json:"customerKey"`
	CardNumber             string `json:"cardNumber"`
	CardExpirationYear     string `json:"cardExpirationYear"`
	CardExpirationMonth    string `json:"cardExpirationMonth"`
	CardPassword           string `json:"cardPassword"`
	CustomerIdentityNumber string `json:"customerIdentityNumber"`
	// TODO (taekyeom) Do we need email or customer name?
}

type tossPaymentCardRegisterResponse struct {
	Mid             string `json:"mid"`
	CustomerKey     string `json:"customerKey"`
	AuthenticatedAt string `json:"authenticatedAt"`
	Method          string `json:"method"`
	BillingKey      string `json:"billingKey"`
	Card            struct {
		Comany    string `json:"company"`
		Number    string `json:"number"`
		CardType  string `json:"cardType"`
		OnwerType string `json:"ownerType"`
	} `json:"card"`
}

type tossPaymentTransactionRequest struct {
	Amount      int    `json:"amount"`
	CustomerKey string `json:"customerKey"`
	OrderId     string `json:"orderId"`
	OrderName   string `json:"orderName"`
}

type tossPaymentObjectResponse struct {
	Version     string `json:"version"`
	PaymentKey  string `json:"paymentKey"`
	Type        string `json:"type"`
	OrderId     string `json:"orderId"`
	OrderName   string `json:"orderName"`
	TotalAmount int    `json:"totalAmount"`
	Status      string `json:"status"`
	Receipt     struct {
		Url string `json:"url"`
	} `json:"receipt"`
}

type tossPaymentTransactionCancelRequest struct {
	CancelReason string `json:"CancelReason"`
	CancelAmount int    `json:"cancelAmount"`
}

func (t tossPaymentService) RegisterCard(ctx context.Context, customerKey string, req request.UserPaymentRegisterRequest) (value.CardPaymentInfo, error) {
	tossPaymentRequest := tossPaymentCardRegisterRequest{
		CustomerKey:            customerKey,
		CardNumber:             req.CardNumber,
		CardExpirationYear:     req.ExpirationYear,
		CardExpirationMonth:    req.ExpirationMonth,
		CustomerIdentityNumber: req.CustomerIdentityNumber,
	}
	resp, err := t.client.R().
		SetBody(tossPaymentRequest).
		SetResult(&tossPaymentCardRegisterResponse{}).
		Post(tossPaymentCardReigstrationPath)

	if err != nil {
		// TODO(taekyeom) Error handling
		return value.CardPaymentInfo{}, fmt.Errorf("%w: error from card registration: %v", value.ErrExternalPayment, err)
	}

	if resp.Error() != nil {
		serviceErr := resp.Error().(*tossPaymentServiceError)
		return value.CardPaymentInfo{}, handleTossPaymentServiceError(serviceErr)
	}

	tossPaymentResp := resp.Result().(*tossPaymentCardRegisterResponse)

	return value.CardPaymentInfo{
		CustomerKey:         tossPaymentResp.CustomerKey,
		CardCompany:         tossPaymentResp.Card.Comany,
		CardNumber:          tossPaymentResp.Card.Number,
		CardExpirationYear:  req.ExpirationYear,
		CardExpirationMonth: req.ExpirationMonth,
		BillingKey:          tossPaymentResp.BillingKey,
	}, nil
}

func (t tossPaymentService) Transaction(ctx context.Context, userPayment entity.UserPayment, payment value.Payment) (value.PaymentResult, error) {
	tossPaymentRequest := tossPaymentTransactionRequest{
		Amount:      payment.Amount,
		CustomerKey: userPayment.Id,
		OrderId:     payment.OrderId,
		OrderName:   payment.OrderName,
	}
	resp, err := t.client.R().
		SetBody(tossPaymentRequest).
		SetResult(&tossPaymentObjectResponse{}).
		SetError(&tossPaymentServiceError{}).
		SetHeader("TossPayments-Test-Code", "INVALID_CARD_EXPIRATION").
		Post(fmt.Sprintf(tossPaymentTransactionPath, userPayment.BillingKey))

	if err != nil {
		// TODO(taekyeom) Error handling
		return value.PaymentResult{}, fmt.Errorf("%w: error while invoking card transaction: %v", value.ErrExternalPayment, err)
	}

	if resp.Error() != nil {
		serviceErr := resp.Error().(*tossPaymentServiceError)
		return value.PaymentResult{}, handleTossPaymentServiceError(serviceErr)
	}

	transactionResp := resp.Result().(*tossPaymentObjectResponse)

	result := value.PaymentResult{
		OrderId:    transactionResp.OrderId,
		PaymentKey: transactionResp.PaymentKey,
		Amount:     transactionResp.TotalAmount,
		OrderName:  transactionResp.OrderName,
		ReceiptUrl: transactionResp.Receipt.Url,
	}

	return result, nil
}

func (t tossPaymentService) CancelTransaction(ctx context.Context, cancel value.PaymentCancel) error {
	tossPaymentRequest := tossPaymentTransactionCancelRequest{
		CancelReason: cancel.Reason,
		CancelAmount: cancel.CancelAmount,
	}

	resp, err := t.client.R().
		SetBody(tossPaymentRequest).
		Post(fmt.Sprintf(tossPaymentCancelTransactionPath, cancel.PaymentKey))
	if err != nil {
		return fmt.Errorf("%w: error from transaction cancellation: %v", value.ErrExternalPayment, err)
	}

	if resp.Error() != nil {
		serviceErr := resp.Error().(*tossPaymentServiceError)
		return handleTossPaymentServiceError(serviceErr)
	}

	return nil
}

func (t tossPaymentService) GetTransactionResult(ctx context.Context, orderId string) (value.PaymentResult, error) {
	resp, err := t.client.R().
		SetResult(&tossPaymentObjectResponse{}).
		SetError(&tossPaymentServiceError{}).
		Get(fmt.Sprintf(tossPaymentGetTransaction, orderId))

	if err != nil {
		return value.PaymentResult{}, fmt.Errorf("%w: error from transaction cancellation: %v", value.ErrExternalPayment, err)
	}

	if resp.Error() != nil {
		serviceErr := resp.Error().(*tossPaymentServiceError)
		return value.PaymentResult{}, handleTossPaymentServiceError(serviceErr)
	}

	transactionResp := resp.Result().(*tossPaymentObjectResponse)

	result := value.PaymentResult{
		OrderId:    transactionResp.OrderId,
		PaymentKey: transactionResp.PaymentKey,
		Amount:     transactionResp.TotalAmount,
		OrderName:  transactionResp.OrderName,
		ReceiptUrl: transactionResp.Receipt.Url,
	}

	return result, nil
}

func NewTossPaymentService(endpoint string, apiKey string) *tossPaymentService {
	client := resty.New().
		SetBaseURL(endpoint).
		SetAuthScheme("Basic").
		SetHeader("Content-Type", "application/json").
		SetError(&tossPaymentServiceError{}).
		SetAuthToken(apiKey)

	return &tossPaymentService{
		client: client,
	}
}

func handleTossPaymentServiceError(serviceError *tossPaymentServiceError) value.TacoError {
	var errCode value.ErrCode
	errMessage := serviceError.Message
	switch serviceError.Code {
	case "DUPLICATED_ORDER_ID":
		errCode = value.ERR_PAYMENT_DUPLICATED_ORDER
	case "INVALID_CARD_EXPIRATION":
		errCode = value.ERR_PAYMENT_INVALID_CARD_EXPIRATION
	case "INVALID_CARD_NUMBER":
		errCode = value.ERR_PAYMENT_INVALID_CARD_NUMBER
	case "INVALID_STOPPED_CARD":
		errCode = value.ERR_PAYMENT_INVALID_STOPPED_CARD
	case "REJECT_ACCOUNT_PAYMENT":
		errCode = value.ERR_PAYMENT_REJECT_ACCOUNT_PAYMENT
	default:
		errCode = value.ERR_EXTERNAL
	}

	return value.NewTacoError(errCode, errMessage)
}

func AsPaymentError(err error, target *value.TacoError) bool {
	if !errors.As(err, target) {
		return false
	}

	if target.Is(value.ErrExternalPayment) ||
		target.Is(value.ErrPaymentDuplicatedOrder) ||
		target.Is(value.ErrPaymentInvalidCardExpiration) ||
		target.Is(value.ErrPaymentInvalidCardNumber) ||
		target.Is(value.ErrPaymentInvalidStoppedCard) ||
		target.Is(value.ErrPaymentRejectAccountPayment) {
		return true
	}

	return false
}
