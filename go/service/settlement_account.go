package service

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
)

type SettlementAccountService interface {
	AuthorizeSettlementAccount(context.Context, entity.Driver, entity.DriverSettlementAccount) (bool, error)
}

type paypleSettlementAccountService struct {
	client      *resty.Client
	customerId  string
	customerKey string
}

func (p paypleSettlementAccountService) AuthorizeSettlementAccount(ctx context.Context, driver entity.Driver,
	settlementAccount entity.DriverSettlementAccount) (bool, error) {
	authResp, err := p.partnerAuthnetication()

	if err != nil {
		return false, err
	}

	request := paypleSettlementAccountAuthorizeRequest{
		CustomerId:            p.customerId,
		CustomerKey:           p.customerKey,
		SubId:                 driver.Id,
		BankCode:              settlementAccount.Bank,
		AccountNum:            settlementAccount.AccountNumber,
		AccountHolderInfoType: "0", // TODO (taekyeom) 법인계좌 지원할 때 하드코드 제거 필요
		AccountholderInfo:     driver.BirthDay,
	}

	resp, err := p.client.R().
		SetAuthScheme(authResp.TokenType).SetAuthToken(authResp.AccessToken).
		SetBody(request).
		SetResult(&paypleSettlementAccountAuthorizeResponse{}).
		Post("inquiry/real_name")

	if err != nil {
		return false, fmt.Errorf("%w: error from back account authorization: %v", value.ErrExternal, err)
	}

	authorizeResp := resp.Result().(*paypleSettlementAccountAuthorizeResponse)
	if !authorizeResp.Success() {
		return false, fmt.Errorf("%w: error from payple transaction: messge: [%s]%s", value.ErrExternal, authorizeResp.Result, authorizeResp.Message)
	}

	fmt.Printf("Result: %+v\n", authorizeResp)

	return authorizeResp.AccountHolderName == driver.FullName(), nil
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
	AccountHolderName string `json:"account_holder_name"`
}

func (p paypleSettlementAccountAuthorizeResponse) Success() bool {
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

type mockSettlementAccountService struct{}

func (m mockSettlementAccountService) AuthorizeSettlementAccount(context.Context, entity.Driver, entity.DriverSettlementAccount) (bool, error) {
	return true, nil
}

func NewMockSettlementAccountService() *mockSettlementAccountService {
	return &mockSettlementAccountService{}
}
