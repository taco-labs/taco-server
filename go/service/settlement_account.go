package service

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/taco-labs/taco/go/domain/value"
)

type SettlementAccountService interface {
	GetSettlementAccount(context.Context, string, string, string, string) (value.SettlementAccount, error)
	TransferRequest(context.Context, value.SettlementTransferRequest) (value.SettlementTransfer, error)
	TransferExecution(context.Context, value.SettlementTransfer) error
}

type paypleSettlementAccountService struct {
	client      *resty.Client
	customerId  string
	customerKey string
}

func (p paypleSettlementAccountService) GetSettlementAccount(ctx context.Context,
	driverId string, accountHolderBirthday, bankCode, bankAccountNumber string) (value.SettlementAccount, error) {
	authResp, err := p.partnerAuthnetication()

	if err != nil {
		return value.SettlementAccount{}, err
	}

	request := paypleSettlementAccountAuthorizeRequest{
		CustomerId:            p.customerId,
		CustomerKey:           p.customerKey,
		SubId:                 driverId,
		BankCode:              bankCode,
		AccountNum:            bankAccountNumber,
		AccountHolderInfoType: "0", // TODO (taekyeom) 법인계좌 지원할 때 하드코드 제거 필요
		AccountholderInfo:     accountHolderBirthday,
	}

	resp, err := p.client.R().
		SetAuthScheme(authResp.TokenType).SetAuthToken(authResp.AccessToken).
		SetBody(request).
		SetResult(&paypleSettlementAccountAuthorizeResponse{}).
		Post("inquiry/real_name")

	if err != nil {
		return value.SettlementAccount{}, fmt.Errorf("%w: error from back account authorization: %v", value.ErrExternal, err)
	}

	authorizeResp := resp.Result().(*paypleSettlementAccountAuthorizeResponse)
	if authorizeResp.UnAuthorized() {
		return value.SettlementAccount{}, fmt.Errorf("%w: Invalid settlement account", value.ErrInvalidOperation)
	}
	if !authorizeResp.Success() {
		return value.SettlementAccount{}, fmt.Errorf("%w: error from payple transaction: messge: [%s]%s", value.ErrExternal, authorizeResp.Result, authorizeResp.Message)
	}

	settlementAccount := value.SettlementAccount{
		BankCode:          bankCode,
		AccountNumber:     bankAccountNumber,
		AccountHolderName: authorizeResp.AccountHolderName,
		BankTransactionId: authorizeResp.BankTransactionId,
	}

	return settlementAccount, nil
}

func (p paypleSettlementAccountService) TransferRequest(ctx context.Context, req value.SettlementTransferRequest) (value.SettlementTransfer, error) {
	authResp, err := p.partnerAuthnetication()

	if err != nil {
		return value.SettlementTransfer{}, err
	}

	request := paypleSettlementTransferRequest{
		CustomerId:        p.customerId,
		CustomerKey:       p.customerKey,
		SubId:             req.DriverId,
		DistinctKey:       req.TransferKey,
		BankTransactionId: req.BankTransactionId,
		TransactionAmount: fmt.Sprint(req.Amount),
		PrintMessage:      req.Message,
	}

	resp, err := p.client.R().
		SetAuthScheme(authResp.TokenType).SetAuthToken(authResp.AccessToken).
		SetBody(request).
		SetResult(&paypleSettlementTransferResponse{}).
		Post("transfer/request")

	if err != nil {
		return value.SettlementTransfer{}, fmt.Errorf("%w: error from settlement transfer: %v", value.ErrExternal, err)
	}

	transferResp := resp.Result().(*paypleSettlementTransferResponse)
	if transferResp.Result == "R0132" {
		return value.SettlementTransfer{}, value.ErrAlreadyExists
	}
	if !transferResp.Success() {
		return value.SettlementTransfer{}, fmt.Errorf("%w: error from payple transfer: messge: [%s]%s", value.ErrExternal, transferResp.Result, transferResp.Message)
	}

	return value.SettlementTransfer{ExecutionKey: transferResp.GroupKey}, nil
}

func (p paypleSettlementAccountService) TransferExecution(ctx context.Context, req value.SettlementTransfer) error {
	authResp, err := p.partnerAuthnetication()

	if err != nil {
		return err
	}

	request := paypleSettlementTransferExeuctionRequest{
		CustomerId:        p.customerId,
		CustomerKey:       p.customerKey,
		GroupKey:          req.ExecutionKey,
		BankTransactionId: "ALL",
		ExecutionType:     "NOW",
	}

	resp, err := p.client.R().
		SetAuthScheme(authResp.TokenType).SetAuthToken(authResp.AccessToken).
		SetBody(request).
		SetResult(&paypleSettlementTransferExeuctionResponse{}).
		Post("transfer/execute")

	if err != nil {
		return fmt.Errorf("%w: error from settlement transfer execution: %v", value.ErrExternal, err)
	}

	transferExecutionResp := resp.Result().(*paypleSettlementTransferExeuctionResponse)
	if !transferExecutionResp.Success() {
		return fmt.Errorf("%w: error from payple transfer: messge: [%s]%s", value.ErrExternal, transferExecutionResp.Result, transferExecutionResp.Message)
	}

	return nil
}

func (p paypleSettlementAccountService) partnerAuthnetication() (*paypleSettlementAccountAuthResponse, error) {
	request := paypleSettlementAccountAuthRequest{
		CustomerId:  p.customerId,
		CustomerKey: p.customerKey,
		Code:        "asdfasdf12", // TODO (taekyeom) generate code?
	}

	resp, err := p.client.R().
		SetBody(request).
		SetResult(&paypleSettlementAccountAuthResponse{}).
		Post("oauth/token")

	if err != nil {
		return nil, fmt.Errorf("%w: error while auth request: %v", value.ErrExternal, err)
	}

	authResp := resp.Result().(*paypleSettlementAccountAuthResponse)
	if !authResp.Success() {
		return nil, fmt.Errorf("%w: error response while authentication: %v", value.ErrExternalPayment, authResp.Message)
	}

	// TODO (taekyeom) handle code match?

	return authResp, nil
}

type paypleSettlementAccountAuthRequest struct {
	CustomerId  string `json:"cst_id"`
	CustomerKey string `json:"custKey"`
	Code        string `json:"code"`
}

type paypleSettlementAccountAuthResponse struct {
	Result      string `json:"result"`
	Message     string `json:"message"`
	Code        string `json:"code"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   string `json:"expires_in"`
}

func (p paypleSettlementAccountAuthResponse) Success() bool {
	return p.Result == "T0000" && p.Message == "처리 성공"
}

type paypleSettlementAccountAuthorizeRequest struct {
	CustomerId            string `json:"cst_id"`
	CustomerKey           string `json:"custKey"`
	SubId                 string `json:"sub_id"`
	BankCode              string `json:"bank_code_std"`
	AccountNum            string `json:"account_num"`
	AccountHolderInfoType string `json:"account_holder_info_type"`
	AccountholderInfo     string `json:"account_holder_info"`
}

type paypleSettlementAccountAuthorizeResponse struct {
	Result            string `json:"result"`
	Message           string `json:"message"`
	CustomerId        string `json:"cst_id"`
	SubId             string `json:"sub_id"`
	BankTransactionId string `json:"billing_tran_id"`
	AccountHolderName string `json:"account_holder_name"`
}

func (p paypleSettlementAccountAuthorizeResponse) Success() bool {
	return p.Result == "A0000" && p.Message == "처리 성공"
}

func (p paypleSettlementAccountAuthorizeResponse) UnAuthorized() bool {
	return p.Result == "N0198"
}

type paypleSettlementTransferRequest struct {
	CustomerId        string `json:"cst_id"`
	CustomerKey       string `json:"custKey"`
	SubId             string `json:"sub_id"`
	DistinctKey       string `json:"distinct_key"`
	BankTransactionId string `json:"billing_tran_id"`
	TransactionAmount string `json:"tran_amt"`
	PrintMessage      string `json:"print_content"`
}

type paypleSettlementTransferResponse struct {
	Result            string `json:"result"`
	Message           string `json:"message"`
	CustomerId        string `json:"cst_id"`
	CustomerKey       string `json:"custKey"`
	SubId             string `json:"sub_id"`
	GroupKey          string `json:"group_key"`
	DistinctKey       string `json:"distinct_key"`
	BankTransactionId string `json:"billing_tran_id"`
	TransactionAmount string `json:"tran_amt"`
	PrintMessage      string `json:"print_content"`
}

func (p paypleSettlementTransferResponse) Success() bool {
	return p.Result == "A0000" && p.Message == "처리 성공"
}

type paypleSettlementTransferExeuctionRequest struct {
	CustomerId        string `json:"cst_id"`
	CustomerKey       string `json:"custKey"`
	GroupKey          string `json:"group_key"`
	BankTransactionId string `json:"billing_tran_id"`
	ExecutionType     string `json:"execute_type"`
}

type mockPaypleSettlementTransferExeuctionRequest struct {
	CustomerId        string `json:"cst_id"`
	CustomerKey       string `json:"custKey"`
	GroupKey          string `json:"group_key"`
	BankTransactionId string `json:"billing_tran_id"`
	ExecutionType     string `json:"execute_type"`
	WebhookUrl        string `json:"webhook_url"`
}

type paypleSettlementTransferExeuctionResponse struct {
	Result      string `json:"result"`
	Message     string `json:"message"`
	CustomerId  string `json:"cst_id"`
	CustomerKey string `json:"custKey"`
}

func (p paypleSettlementTransferExeuctionResponse) Success() bool {
	return p.Result == "A0000" && p.Message == "처리 성공"
}

func NewPaypleSettlemtnAccountService(serviceEndpoint, customerId, customerKey string) *paypleSettlementAccountService {
	client := resty.New().
		SetBaseURL(serviceEndpoint).
		SetHeaders(map[string]string{
			"Content-Type":  "application/json",
			"Cache-Control": "no-cache",
		})

	return &paypleSettlementAccountService{
		client:      client,
		customerId:  customerId,
		customerKey: customerKey,
	}
}

type mockSettlementAccountService struct {
	inner      *paypleSettlementAccountService
	webhookUrl string
}

func (m mockSettlementAccountService) GetSettlementAccount(ctx context.Context,
	driverId, accountHolderBirthday, bankCode, bankAccountNumber string) (value.SettlementAccount, error) {
	return m.inner.GetSettlementAccount(ctx, driverId, accountHolderBirthday, bankCode, bankAccountNumber)
}

func (m mockSettlementAccountService) TransferRequest(ctx context.Context, req value.SettlementTransferRequest) (value.SettlementTransfer, error) {
	// For test purpose..
	req.Amount = 1000
	return m.inner.TransferRequest(ctx, req)
}

func (m mockSettlementAccountService) TransferExecution(ctx context.Context, req value.SettlementTransfer) error {
	authResp, err := m.inner.partnerAuthnetication()

	if err != nil {
		return err
	}

	request := mockPaypleSettlementTransferExeuctionRequest{
		CustomerId:        m.inner.customerId,
		CustomerKey:       m.inner.customerKey,
		GroupKey:          req.ExecutionKey,
		BankTransactionId: "ALL",
		ExecutionType:     "NOW",
		WebhookUrl:        m.webhookUrl,
	}

	resp, err := m.inner.client.R().
		SetAuthScheme(authResp.TokenType).SetAuthToken(authResp.AccessToken).
		SetBody(request).
		SetResult(&paypleSettlementTransferExeuctionResponse{}).
		Post("transfer/execute")

	if err != nil {
		return fmt.Errorf("%w: error from settlement transfer execution: %v", value.ErrExternal, err)
	}

	transferExecutionResp := resp.Result().(*paypleSettlementTransferExeuctionResponse)
	if !transferExecutionResp.Success() {
		return fmt.Errorf("%w: error from payple transfer: messge: [%s]%s", value.ErrExternal, transferExecutionResp.Result, transferExecutionResp.Message)
	}

	return nil
}

func NewMockSettlementAccountService(serviceEndpoint, customerId, customerKey, mockWebhookUrl string) *mockSettlementAccountService {
	return &mockSettlementAccountService{
		inner:      NewPaypleSettlemtnAccountService(serviceEndpoint, customerId, customerKey),
		webhookUrl: mockWebhookUrl,
	}
}
