package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
)

type PaymentService interface {
	GetCardRegistrationRequestParam(context.Context, int, entity.User) (value.PaymentRegistrationRequestParam, error)
	DeleteCard(context.Context, string) error
	Transaction(context.Context, entity.UserPayment, value.Payment) (value.PaymentResult, error) // TODO(taekyeom) 결제 기록 별도 보관 필요
	CancelTransaction(context.Context, value.PaymentCancel) error
	GetTransactionResult(context.Context, string) (value.PaymentResult, error)
}

type payplePaymentService struct {
	client      *resty.Client
	customerId  string
	customerKey string
}

type paypleCardRegistrationRequestParamResponse struct {
	ServerName  string `json:"server_name"`
	Result      string `json:"result"`
	ResultMsg   string `json:"result_msg"`
	CustomerId  string `json:"cst_id"`
	CustomerKey string `json:"custKey"`
	AuthKey     string `json:"AuthKey"`
	ReturnUrl   string `json:"return_url"`
}

type payplePartnerAuthenticationRequest struct {
	CustomerId  string `json:"cst_id"`
	CustomerKey string `json:"custKey"`
	PayType     string `json:"PCD_PAY_TYPE"`
	RestFlag    string `json:"PCD_SIMPLE_FLAG"`
}

type payplePartnerAuthenticationResponse struct {
	ServerName  string `json:"server_name"`
	Result      string `json:"result"`
	ResultMsg   string `json:"result_msg"`
	CustomerId  string `json:"cst_id"`
	CustomerKey string `json:"custKey"`
	AuthKey     string `json:"AuthKey"`
	PayUrl      string `json:"PCD_PAY_URL"`
	ReturnUrl   string `json:"return_url"`
}

func (p payplePartnerAuthenticationResponse) Success() bool {
	return p.Result == "success"
}

type paypleTransactionRequest struct {
	CustomerId  string `json:"PCD_CST_ID"`
	CustomerKey string `json:"PCD_CUST_KEY"`
	AuthKey     string `json:"PCD_AUTH_KEY"`
	PayType     string `json:"PCD_PAY_TYPE"`
	BillingKey  string `json:"PCD_PAYER_ID"`
	OrderName   string `json:"PCD_PAY_GOODS"`
	RestFlag    string `json:"PCD_SIMPLE_FLAG"`
	Amount      int    `json:"PCD_PAY_TOTAL"`
	OrderId     string `json:"PCD_PAY_OID"`
}

type paypleTransactionResponse struct {
	Result     string `json:"PCD_PAY_RST"`
	ResultCode string `json:"PCD_PAY_CODE"`
	ResultMsg  string `json:"PCD_PAY_MSG"`
	OrderId    string `json:"PCD_PAY_OID"`
	PaymentKey string `json:"PCD_PAY_CARDTRADENUM"`
	ReceiptUrl string `json:"PCD_PAY_CARDRECEIPT"`
}

func (p paypleTransactionResponse) Success() bool {
	return p.Result == "success"
}

type paypleGetTransactionRequest struct {
}

type paypleDeletePaymentRequest struct {
	CustomerId  string `json:"PCD_CST_ID"`
	CustomerKey string `json:"PCD_CUST_KEY"`
	AuthKey     string `json:"PCD_AUTH_KEY"`
	BillingKey  string `json:"PCD_PAYER_ID"`
}

type paypleDeletePaymentResponse struct {
	Result  string `json:"PCD_PAY_RST"`
	Code    string `json:"PCD_PAY_CODE"`
	Message string `json:"PCD_PAY_MSG"`
}

func (p paypleDeletePaymentResponse) Success() bool {
	return p.Result == "success" && p.Code == "PUER0000"
}

func (p payplePaymentService) GetCardRegistrationRequestParam(ctx context.Context, requestId int, user entity.User) (value.PaymentRegistrationRequestParam, error) {
	request := struct {
		CustomerId  string `json:"cst_id"`
		CustomerKey string `json:"custKey"`
	}{
		p.customerId,
		p.customerKey,
	}

	resp, err := p.client.R().
		SetBody(request).
		SetResult(&paypleCardRegistrationRequestParamResponse{}).
		Post("php/auth.php")

	if err != nil {
		return value.PaymentRegistrationRequestParam{},
			fmt.Errorf("%w: error from payple card registration auth request: %v", value.ErrExternalPayment, err)
	}

	// TODO (taekyeom) handle success check?
	authResp := resp.Result().(*paypleCardRegistrationRequestParamResponse)

	return value.PaymentRegistrationRequestParam{
		RequestId:       requestId,
		AuthKey:         authResp.AuthKey,
		RegistrationUrl: authResp.ReturnUrl,
		UserPhone:       user.Phone,
	}, nil
}

func (p payplePaymentService) Transaction(ctx context.Context, userPayment entity.UserPayment, payment value.Payment) (value.PaymentResult, error) {
	authResp, err := p.partnerAuthentication(ctx)
	if err != nil {
		return value.PaymentResult{}, err
	}

	request := paypleTransactionRequest{
		CustomerId:  authResp.CustomerId,
		CustomerKey: authResp.CustomerKey,
		AuthKey:     authResp.AuthKey,
		PayType:     "card",
		BillingKey:  userPayment.BillingKey,
		OrderName:   payment.OrderName,
		RestFlag:    "Y",
		Amount:      payment.Amount,
		OrderId:     payment.OrderId,
	}

	resp, err := p.client.R().
		SetBody(request).
		SetResult(&paypleTransactionResponse{}).
		Post(authResp.PayUrl)

	if err != nil {
		return value.PaymentResult{}, fmt.Errorf("%w: error from payple transaction request: %v", value.ErrExternalPayment, err)
	}

	transactionResp := resp.Result().(*paypleTransactionResponse)
	if !transactionResp.Success() {
		return value.PaymentResult{}, fmt.Errorf("%w: error from payple transaction: messge: [%s]%s", value.ErrExternalPayment, transactionResp.ResultCode, transactionResp.ResultMsg)
	}

	return value.PaymentResult{
		OrderId:    transactionResp.OrderId,
		PaymentKey: transactionResp.PaymentKey,
		Amount:     payment.Amount,
		OrderName:  payment.OrderName,
		ReceiptUrl: transactionResp.ReceiptUrl,
	}, nil
}

func (p payplePaymentService) DeleteCard(ctx context.Context, billingKey string) error {
	request := struct {
		CustomerId  string `json:"cst_id"`
		CustomerKey string `json:"custKey"`
		PayWork     string `json:"PCD_PAY_WORK"`
	}{
		p.customerId,
		p.customerKey,
		"PUSERDEL",
	}

	resp, err := p.client.R().
		SetBody(request).
		SetResult(&payplePartnerAuthenticationResponse{}).
		Post("php/auth.php")

	if err != nil {
		return fmt.Errorf("%w: error while auth request: %v", value.ErrExternalPayment, err)
	}

	authResp := resp.Result().(*payplePartnerAuthenticationResponse)
	if !authResp.Success() {
		return fmt.Errorf("%w: error response while authentication: %v", value.ErrExternalPayment, authResp.ResultMsg)
	}

	deleteRequest := paypleDeletePaymentRequest{
		authResp.CustomerId,
		authResp.CustomerKey,
		authResp.AuthKey,
		billingKey,
	}

	resp, err = p.client.R().
		SetBody(deleteRequest).
		SetResult(&paypleDeletePaymentResponse{}).
		Post(authResp.PayUrl)

	if err != nil {
		return fmt.Errorf("%w: error while delete request: %v", value.ErrExternalPayment, err)
	}

	deleteResp := resp.Result().(*paypleDeletePaymentResponse)
	if !deleteResp.Success() {
		return fmt.Errorf("%w: error response while delete payment: %v", value.ErrExternalPayment, deleteResp.Message)
	}

	return nil
}

func (p payplePaymentService) CancelTransaction(context.Context, value.PaymentCancel) error {
	return value.ErrUnsupported
}

func (p payplePaymentService) GetTransactionResult(context.Context, string) (value.PaymentResult, error) {
	return value.PaymentResult{}, value.ErrUnsupported
}

func (p payplePaymentService) partnerAuthentication(context.Context) (*payplePartnerAuthenticationResponse, error) {
	request := payplePartnerAuthenticationRequest{
		CustomerId:  p.customerId,
		CustomerKey: p.customerKey,
		PayType:     "card",
		RestFlag:    "Y",
	}

	resp, err := p.client.R().
		SetBody(request).
		SetResult(&payplePartnerAuthenticationResponse{}).
		Post("php/auth.php")

	if err != nil {
		return nil, fmt.Errorf("%w: error while auth request: %v", value.ErrExternalPayment, err)
	}

	authResp := resp.Result().(*payplePartnerAuthenticationResponse)
	if !authResp.Success() {
		return nil, fmt.Errorf("%w: error response while authentication: %v", value.ErrExternalPayment, authResp.ResultMsg)
	}

	return authResp, nil
}

func NewPayplePaymentService(serviceEndpoint, refererHost, customerId, customerKey string) *payplePaymentService {
	client := resty.New().
		SetBaseURL(serviceEndpoint).
		SetHeaders(map[string]string{
			"Content-Type":  "application/json",
			"Cache-Control": "no-cache",
			"Referer":       refererHost,
		})

	return &payplePaymentService{
		client:      client,
		customerId:  customerId,
		customerKey: customerKey,
	}
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
