package payple

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"text/template"

	"github.com/labstack/echo/v4"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/response"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/server"
	"github.com/taco-labs/taco/go/utils"
)

type paypleExtension struct {
	app struct {
		userApp userApp
	}
	domain   string
	renderer *paypleTemplate
}

type userApp interface {
	GetCardRegistrationRequestParam(context.Context, string) (value.PaymentRegistrationRequestParam, error)
	RegistrationCallback(context.Context, request.PaymentRegistrationCallbackRequest) (entity.UserPayment, error)
}

type paypleResultRequest struct {
	Result        string `form:"PCD_PAY_RST"`
	RequestId     string `form:"PCD_PAYER_NO"`
	ResultMessage string `form:"PCD_PAY_CODE"`
	BillingKey    string `form:"PCD_PAYER_ID"`
	CardCompany   string `form:"PCD_PAY_CARDNAME"`
	CardNumber    string `form:"PCD_PAY_CARDNUM"`
}

func (p paypleResultRequest) Success() bool {
	return p.Result == "success"
}

func (p paypleExtension) RegistCardPayment(e echo.Context) error {
	ctx := e.Request().Context()

	userId := utils.GetUserId(ctx)

	resp, err := p.app.userApp.GetCardRegistrationRequestParam(ctx, userId)
	if err != nil {
		return server.ToResponse(err)
	}

	params := map[string]interface{}{
		"authKey":   resp.AuthKey,
		"payUrl":    resp.RegistrationUrl,
		"resultUrl": fmt.Sprintf("%s/payment/payple/result_callback", p.domain),
		"userPhone": resp.UserPhone,
		"requestId": resp.RequestId,
	}

	return e.Render(http.StatusOK, "payple.html", params)
}

func (p paypleExtension) RegisterCardPaymentResultCallback(e echo.Context) error {
	ctx := e.Request().Context()

	body := paypleResultRequest{}
	if err := e.Bind(&body); err != nil {
		return err
	}
	if !body.Success() {
		return server.ToResponse(value.NewTacoError(value.ERR_EXTERNAL_PAYMENT, body.ResultMessage))
	}

	requestId, _ := strconv.Atoi(body.RequestId)
	req := request.PaymentRegistrationCallbackRequest{
		RequestId:   requestId,
		BillingKey:  body.BillingKey,
		CardCompany: body.CardCompany,
		CardNumber:  body.CardNumber,
	}

	userPayment, err := p.app.userApp.RegistrationCallback(ctx, req)
	if err != nil {
		return server.ToResponse(err)
	}

	return e.JSON(http.StatusOK, response.UserPaymentToResponse(userPayment))
}

func (p paypleExtension) Apply(e *echo.Echo) {
	e.GET("/payment/payple/register", p.RegistCardPayment)
	e.POST("/payment/payple/result_callback", p.RegisterCardPaymentResultCallback)
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
