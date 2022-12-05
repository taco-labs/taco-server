package payple

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/server"
)

type paypleExtension struct {
	app struct {
		driversettlement driversettlementApp
	}
}

type driversettlementApp interface {
	DriverSettlementTransferSuccessCallback(ctx context.Context, req request.DriverSettlementTransferSuccessCallbackRequest) error
	DriverSettlementTransferFailureCallback(ctx context.Context, req request.DriverSettlementTransferFailureCallbackRequest) error
}

func (p paypleExtension) SettlementTransferCallback(e echo.Context) error {
	ctx := e.Request().Context()

	req := paypleSettlementTransferCallbackRequest{}

	if err := e.Bind(&req); err != nil {
		return err
	}

	var err error
	if req.Success() {
		err = p.app.driversettlement.DriverSettlementTransferSuccessCallback(ctx, request.DriverSettlementTransferSuccessCallbackRequest{
			DriverId: req.SubId,
		})
	} else {
		err = p.app.driversettlement.DriverSettlementTransferFailureCallback(ctx, request.DriverSettlementTransferFailureCallbackRequest{
			DriverId:       req.SubId,
			FailureMessage: req.Error(),
		})
	}

	if err != nil {
		return server.ToResponse(e, err)
	}

	return e.JSON(http.StatusOK, struct{}{})
}

func (p paypleExtension) Apply(e *echo.Echo) {
	e.POST("/payment/payple/settlement_callback", p.SettlementTransferCallback)
}

type paypleSettlementTransferCallbackRequest struct {
	Result  string `json:"result"`
	Message string `json:"message"`
	SubId   string `json:"sub_id"`
}

func (p paypleSettlementTransferCallbackRequest) Success() bool {
	return p.Result == "A0000" && p.Message == "처리 성공"
}

func (p paypleSettlementTransferCallbackRequest) Error() string {
	return fmt.Sprintf("[%s]%s", p.Result, p.Message)
}

// {
//    "result":"A0000",
//    "message":"처리 성공",
//    "cst_id":"test",
//    "sub_id":"test12385610",
//    "group_key":"Q0RSSkYzWUI3...",
//    "billing_tran_id":"6fen3g2m-j9hb-...",
//    "api_tran_id":"ohr8ps3m-j5x...",
//    "api_tran_dtm":"20211025142315647",
//    "bank_tran_id":"M202112389U142315646",
//    "bank_tran_date":"20211025",
//    "bank_rsp_code":"000",
//    "bank_code_std":"020",
//    "bank_code_sub":"0000000",
//    "bank_name":"우리은행",
//    "account_num":"1234567890123456",
//    "account_num_masked":"1234567890123***",
//    "account_holder_name":"홍길동",
//    "print_content":"정산테스트",
//    "tran_amt":"1000"
// }
