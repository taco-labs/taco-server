package service

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/utils"
)

const (
	tossPaymentCardReigstrationPath = "v1/billing/authorizations/card"
	tossPaymentTransactionPath      = "v1/billing/%s"
)

type CardPaymentService interface {
	RegisterCard(context.Context, entity.User, request.UserPaymentRegisterRequest) (entity.UserPayment, error)
	Transaction(context.Context, entity.UserPayment, value.Payment) error // TODO(taekyeom) 결제 기록 별도 보관 필요
}

type tossPaymentService struct {
	client *resty.Client
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

type tossPaymentTransactionResponse struct {
	// TODO (taekyeom) Fill...
}

func (t tossPaymentService) RegisterCard(ctx context.Context, user entity.User, req request.UserPaymentRegisterRequest) (entity.UserPayment, error) {
	tossPaymentRequest := tossPaymentCardRegisterRequest{
		CustomerKey:            utils.MustNewUUID(), // TODO (taekyeom) inject id from outside
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
		return entity.UserPayment{}, fmt.Errorf("%w: error from card registration: %v", value.ErrExternal, err)
	}

	tossPaymentResp := resp.Result().(*tossPaymentCardRegisterResponse)

	return entity.UserPayment{
		Id:                  tossPaymentResp.CustomerKey,
		UserId:              user.Id,
		Name:                req.Name,
		CardCompany:         tossPaymentResp.Card.Comany,
		RedactedCardNumber:  tossPaymentResp.Card.Number,
		CardExpirationYear:  req.ExpirationYear,
		CardExpirationMonth: req.ExpirationMonth,
		BillingKey:          tossPaymentResp.BillingKey,
		DefaultPayment:      req.DefaultPayment,
		CreateTime:          time.Now().UTC(),
	}, nil
}

func (t tossPaymentService) Transaction(ctx context.Context, userPayment entity.UserPayment, payment value.Payment) error {
	tossPaymentRequest := tossPaymentTransactionRequest{
		Amount:      payment.Amount,
		CustomerKey: userPayment.Id,
		OrderId:     payment.OrderId,
		OrderName:   payment.OrderName,
	}
	_, err := t.client.R().
		SetBody(tossPaymentRequest).
		SetResult(&tossPaymentTransactionResponse{}).
		Post(fmt.Sprintf(tossPaymentTransactionPath, userPayment.BillingKey))
	if err != nil {
		// TODO(taekyeom) Error handling
		return fmt.Errorf("%w: error from card transaction: %v", value.ErrExternal, err)
	}

	// TODO (taekyoem) handle response from transaction
	return nil
}

func NewTossPaymentService(endpoint string, apiKey string) tossPaymentService {
	client := resty.New().
		SetBaseURL(endpoint).
		SetAuthScheme("Basic").
		SetHeader("Content-Type", "application/json").
		SetAuthToken(apiKey)

	return tossPaymentService{
		client: client,
	}
}
