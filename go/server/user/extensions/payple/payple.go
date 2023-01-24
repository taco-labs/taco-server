package payple

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/server"
	"github.com/taco-labs/taco/go/utils"
	"go.uber.org/zap"
)

type PaypleTimeFormat time.Time

const (
	paypleTimeLayout = "20060102150405"
)

func (p *PaypleTimeFormat) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")

	loc, _ := time.LoadLocation("Asia/Seoul")

	if s == "null" || s == "" {
		*p = PaypleTimeFormat(time.Time{})
		return
	}
	t, err := time.ParseInLocation(paypleTimeLayout, s, loc)
	*p = PaypleTimeFormat(t)
	return
}

type paypleExtension struct {
	app struct {
		paymentApp paymentApp
	}
	domain   string
	renderer *paypleTemplate
	env      string
}

type paymentApp interface {
	GetCardRegistrationRequestParam(context.Context, string) (value.PaymentRegistrationRequestParam, error)
	RegistrationCallback(context.Context, request.PaymentRegistrationCallbackRequest) (entity.UserPayment, error)
	PaymentTransactionSuccessCallback(ctx context.Context, req request.PaymentTransactionSuccessCallbackRequest) error
	PaymentTransactionFailCallback(ctx context.Context, req request.PaymentTransactionFailCallbackRequest) error
}

type paypleResultRequest struct {
	Result      string `form:"PCD_PAY_RST"`
	Code        string `form:"PCD_PAY_CODE"`
	Message     string `form:"PCD_PAY_MSG"`
	RequestId   string `form:"PCD_PAYER_NO"`
	BillingKey  string `form:"PCD_PAYER_ID"`
	CardCompany string `form:"PCD_PAY_CARDNAME"`
	CardNumber  string `form:"PCD_PAY_CARDNUM"`
}

func (p paypleResultRequest) Success() bool {
	return p.Result == "success" && p.Code == "0000"
}

func (p paypleResultRequest) Cancel() bool {
	return p.Message == "결제를 종료하였습니다."
}

type paypleTransactionResultCallbackRequest struct {
	Result     string           `json:"PCD_PAY_RST"`
	Code       string           `json:"PCD_PAY_CODE"`
	Message    string           `json:"PCD_PAY_MSG"`
	OrderId    string           `json:"PCD_PAY_OID"`
	ReceiptUrl string           `json:"PCD_PAY_CARDRECEIPT"`
	PayTime    PaypleTimeFormat `json:"PCD_PAY_TIME"`
}

func (p paypleTransactionResultCallbackRequest) Success() bool {
	return p.Result == "success" && p.Code == "SPCD0000"
}

func (p paypleTransactionResultCallbackRequest) Cancel() bool {
	return strings.HasPrefix(p.Code, "PAYC")
}

func (p paypleExtension) RegistCardPayment(e echo.Context) error {
	ctx := e.Request().Context()

	userId := utils.GetUserId(ctx)

	resp, err := p.app.paymentApp.GetCardRegistrationRequestParam(ctx, userId)
	if err != nil {
		return server.ToResponse(e, err)
	}

	params := map[string]interface{}{
		"authKey":   resp.AuthKey,
		"payUrl":    resp.RegistrationUrl,
		"resultUrl": fmt.Sprintf("%s/payment/payple/result_callback", p.domain),
		"userPhone": resp.UserPhone,
		"requestId": resp.RequestId,
		"env":       p.env,
	}

	return e.Render(http.StatusOK, "payple_register.html", params)
}

func (p paypleExtension) RegisterCardPaymentResultCallback(e echo.Context) error {
	ctx := e.Request().Context()
	logger := utils.GetLogger(ctx)

	body := paypleResultRequest{}
	if err := e.Bind(&body); err != nil {
		logger.Error("server.user.extension.payple: Error while parse payple result request", zap.Error(err))
		return e.Redirect(http.StatusSeeOther, fmt.Sprintf("%s/payment/payple/register_failure", p.domain))
	}
	if body.Cancel() {
		return e.Redirect(http.StatusSeeOther, fmt.Sprintf("%s/payment/payple/register_cancel", p.domain))
	}
	if !body.Success() {
		logger.Error("server.user.extension.payple: Error from payple card registration result",
			zap.String("errCode", body.Code),
			zap.String("errMessage", body.Message),
		)
		return e.Redirect(http.StatusSeeOther, fmt.Sprintf("%s/payment/payple/register_failure", p.domain))
	}

	requestId, _ := strconv.Atoi(body.RequestId)
	req := request.PaymentRegistrationCallbackRequest{
		RequestId:   requestId,
		BillingKey:  body.BillingKey,
		CardCompany: body.CardCompany,
		CardNumber:  body.CardNumber,
	}

	_, err := p.app.paymentApp.RegistrationCallback(ctx, req)
	if err != nil {
		logger.Error("server.user.extension.payple: Error while registering card", zap.Error(err))
		return e.Redirect(http.StatusSeeOther, fmt.Sprintf("%s/payment/payple/register_failure", p.domain))
	}

	return e.Redirect(http.StatusSeeOther, fmt.Sprintf("%s/payment/payple/register_success", p.domain))
}

func (p paypleExtension) RegisterCardPaymentSuccess(e echo.Context) error {
	return e.Render(http.StatusOK, "payple_register_success.html", map[string]interface{}{})
}

func (p paypleExtension) RegisterCardPaymentFailure(e echo.Context) error {
	return e.Render(http.StatusOK, "payple_register_failure.html", map[string]interface{}{})
}

func (p paypleExtension) RegisterCardPaymentCancel(e echo.Context) error {
	return e.Render(http.StatusOK, "payple_register_cancel.html", map[string]interface{}{})
}

func (p paypleExtension) TransactionResultCallback(e echo.Context) error {
	ctx := e.Request().Context()

	req := paypleTransactionResultCallbackRequest{}

	if err := e.Bind(&req); err != nil {
		return server.ToResponse(e, err)
	}

	var err error
	if req.Success() {
		err = p.app.paymentApp.PaymentTransactionSuccessCallback(ctx, request.PaymentTransactionSuccessCallbackRequest{
			OrderId:    req.OrderId,
			PaymentKey: req.OrderId,
			ReceiptUrl: req.ReceiptUrl,
			CreateTime: time.Time(req.PayTime),
		})
	}
	if !req.Success() && !req.Cancel() {
		err = p.app.paymentApp.PaymentTransactionFailCallback(ctx, request.PaymentTransactionFailCallbackRequest{
			OrderId:       req.OrderId,
			FailureCode:   req.Code,
			FailureReason: req.Message,
		})
	}

	// TODO (taekyeom) 결제 취소 건 대응 필요

	if err != nil {
		return server.ToResponse(e, err)
	}

	return e.JSON(http.StatusOK, struct{}{})
}

func (p paypleExtension) Apply(e *echo.Echo) {
	e.GET("/payment/payple/register", p.RegistCardPayment)
	e.POST("/payment/payple/result_callback", p.RegisterCardPaymentResultCallback)
	e.GET("/payment/payple/register_success", p.RegisterCardPaymentSuccess)
	e.GET("/payment/payple/register_failure", p.RegisterCardPaymentFailure)
	e.GET("/payment/payple/register_cancel", p.RegisterCardPaymentCancel)
	e.POST("/payment/payple/transaction_callback", p.TransactionResultCallback)
	e.Renderer = p.renderer
}

func NewPaypleExtension(opts ...paypleExtensionOption) (*paypleExtension, error) {
	templates, err := template.ParseGlob("go/templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("error while initialize templates: %w", err)
	}

	extension := &paypleExtension{
		renderer: &paypleTemplate{templates: templates},
	}

	for _, opt := range opts {
		opt(extension)
	}

	return extension, extension.validate()
}

type paypleTemplate struct {
	templates *template.Template
}

func (p *paypleTemplate) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return p.templates.ExecuteTemplate(w, name, data)
}
