package driver

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/response"
	"github.com/taco-labs/taco/go/server"
)

type driverApp interface {
	SmsVerificationRequest(context.Context, request.SmsVerificationRequest) (entity.SmsVerification, error)
	SmsSignin(context.Context, request.SmsSigninRequest) (entity.Driver, string, error)
	Signup(context.Context, request.DriverSignupRequest) (entity.Driver, string, error)
}

func (d driverServer) SmsVerificationRequest(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.SmsVerificationRequest{}

	if err := e.Bind(&req); err != nil {
		return err
	}

	smsVerification, err := d.app.driver.SmsVerificationRequest(ctx, req)
	if err != nil {
		return server.ToResponse(err)
	}

	resp := response.SmsVerificationRequestResponse{
		Id:         smsVerification.Id,
		ExpireTime: smsVerification.ExpireTime,
	}

	return e.JSON(http.StatusOK, resp)
}

func (d driverServer) SmsSignin(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.SmsSigninRequest{}

	if err := e.Bind(&req); err != nil {
		return err
	}

	driver, token, err := d.app.driver.SmsSignin(ctx, req)

	if err != nil {
		return server.ToResponse(err)
	}

	resp := response.DriverSignupResponse{
		Token:  token,
		Driver: response.DriverToResponse(driver),
	}

	return e.JSON(http.StatusOK, resp)
}

func (d driverServer) Signup(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.DriverSignupRequest{}

	if err := e.Bind(&req); err != nil {
		return err
	}

	driver, token, err := d.app.driver.Signup(ctx, req)

	if err != nil {
		return server.ToResponse(err)
	}

	resp := response.DriverSignupResponse{
		Token:  token,
		Driver: response.DriverToResponse(driver),
	}

	return e.JSON(http.StatusOK, resp)
}
