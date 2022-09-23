package user

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/response"
)

type UserApp interface {
	SmsVerificationRequest(context.Context, request.SmsVerificationRequest) (entity.SmsVerification, error)
	SmsSignin(context.Context, request.SmsSigninRequest) (entity.User, string, error)
	Signup(context.Context, request.UserSignupRequest) (entity.User, string, error)
	GetUser(context.Context, string) (entity.User, error)
	UpdateUser(context.Context, request.UserUpdateRequest) (entity.User, error)
	ListCardPayment(ctx context.Context, userId string) ([]entity.UserPayment, error)
	RegisterCardPayment(ctx context.Context, req request.UserPaymentRegisterRequest) (entity.UserPayment, error)
	DeleteCardPayment(ctx context.Context, userPaymentId string) error
	UpdateDefaultPayment(ctx context.Context, req request.DefaultPaymentUpdateRequest) error
}

func (u userServer) SmsVerificationRequest(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.SmsVerificationRequest{}

	if err := e.Bind(&req); err != nil {
		return e.String(http.StatusBadRequest, "bind error")
	}

	smsVerification, err := u.app.user.SmsVerificationRequest(ctx, req)
	if err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}

	resp := response.SmsVerificationRequestResponse{
		Id:         smsVerification.Id,
		ExpireTime: smsVerification.ExpireTime,
	}

	return e.JSON(http.StatusOK, resp)
}

func (u userServer) SmsSingin(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.SmsSigninRequest{}

	if err := e.Bind(&req); err != nil {
		return e.String(http.StatusBadRequest, "bind error")
	}

	user, token, err := u.app.user.SmsSignin(ctx, req)

	resp := response.UserSignupResponse{
		Token: token,
		User:  response.UserToResponse(user),
	}

	if err != nil {
		// TODO(taekyeom) Error handle
		return e.String(http.StatusBadRequest, err.Error())
	}

	return e.JSON(http.StatusOK, resp)
}

func (u userServer) Signup(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.UserSignupRequest{}

	if err := e.Bind(&req); err != nil {
		// TODO(taekyeom) Error handle
		return e.String(http.StatusBadRequest, "bind error 1")
	}

	user, token, err := u.app.user.Signup(ctx, req)

	resp := response.UserSignupResponse{
		Token: token,
		User:  response.UserToResponse(user),
	}

	if err != nil {
		// TODO(taekyeom) Error handle
		return e.String(http.StatusBadRequest, err.Error())
	}

	return e.JSON(http.StatusOK, resp)
}

func (u userServer) GetUser(e echo.Context) error {
	ctx := e.Request().Context()
	userId := e.Param("userId")
	user, err := u.app.user.GetUser(ctx, userId)
	if err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}
	return e.JSON(http.StatusOK, response.UserToResponse(user))
}

func (u userServer) UpdateUser(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.UserUpdateRequest{}

	if err := e.Bind(&req); err != nil {
		// TODO(taekyeom) Error handle
		return e.String(http.StatusBadRequest, "bind error 1")
	}

	user, err := u.app.user.UpdateUser(ctx, req)
	if err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}

	return e.JSON(http.StatusOK, response.UserToResponse(user))
}

func (u userServer) ListCardPayment(e echo.Context) error {
	ctx := e.Request().Context()
	userId := e.Param("userId")
	cardPayments, err := u.app.user.ListCardPayment(ctx, userId)
	if err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}
	return e.JSON(http.StatusOK, response.ListCardPaymentResponse{
		Payments: response.UserPaymentsToResponse(cardPayments),
	})
}

func (u userServer) RegisterCardPayment(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.UserPaymentRegisterRequest{}
	if err := e.Bind(&req); err != nil {
		// TODO(taekyeom) Error handle
		return e.String(http.StatusBadRequest, fmt.Errorf("bind error: %v", err).Error())
	}

	cardPayment, err := u.app.user.RegisterCardPayment(ctx, req)
	if err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}

	return e.JSON(http.StatusOK, response.UserPaymentToResponse(cardPayment))
}

func (u userServer) DeleteCardPayment(e echo.Context) error {
	ctx := e.Request().Context()

	paymentId := e.Param("paymentId")
	err := u.app.user.DeleteCardPayment(ctx, paymentId)
	if err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}
	return e.JSON(http.StatusOK, response.DeleteCardPaymentResponse{
		PaymentId: paymentId,
	})
}

func (u userServer) UpdateDefaultPayment(e echo.Context) error {
	return nil
}
