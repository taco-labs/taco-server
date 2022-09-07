package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ktk1012/taco/go/domain/entity"
	"github.com/ktk1012/taco/go/domain/request"
	"github.com/ktk1012/taco/go/domain/response"
	"github.com/ktk1012/taco/go/service"
	"github.com/labstack/echo/v4"
)

type UserApp interface {
	Signup(context.Context, request.UserSignupRequest) (entity.User, string, error)
	GetUser(context.Context, string) (entity.User, error)
	ListCardPayment(ctx context.Context, userId string) ([]entity.UserPayment, error)
	RegisterCardPayment(ctx context.Context, req request.UserPaymentRegisterRequest) (entity.UserPayment, error)
	DeleteCardPayment(ctx context.Context, userPaymentId string) error
	UpdateDefaultPayment(ctx context.Context, req request.DefaultPaymentUpdateRequest) (entity.User, error)
}

func (u userServer) Signup(e echo.Context) error {
	ctx := e.Request().Context()

	// TODO(taekyeom) Remove mock
	type mockReq struct {
		request.UserSignupRequest
		request.MockUserIdentity
	}

	req := mockReq{}

	if err := e.Bind(&req); err != nil {
		// TODO(taekyeom) Error handle
		return e.String(http.StatusBadRequest, "bind error 1")
	}

	ctx = service.SetMockIdentity(ctx, req.MockUserIdentity)

	user, token, err := u.app.user.Signup(ctx, req.UserSignupRequest)

	resp := response.UserSignupResponse{
		Token: token,
		User:  response.UserResponseFromUser(user),
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
	return e.JSON(http.StatusOK, user)
}

func (u userServer) ListCardPayment(e echo.Context) error {
	ctx := e.Request().Context()
	userId := e.Param("userId")
	cardPayments, err := u.app.user.ListCardPayment(ctx, userId)
	if err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}
	return e.JSON(http.StatusOK, cardPayments)
}

func (u userServer) RegisterCardPayment(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.UserPaymentRegisterRequest{}
	if err := e.Bind(&req); err != nil {
		// TODO(taekyeom) Error handle
		return e.String(http.StatusBadRequest, fmt.Errorf("bind error: %v", err).Error())
	}

	fmt.Printf("Test: %v\n", req)

	cardPayment, err := u.app.user.RegisterCardPayment(ctx, req)
	if err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}

	fmt.Printf("Test Done: %v\n", cardPayment)

	return e.JSON(http.StatusOK, cardPayment)
}

func (u userServer) DeleteCardPayment(e echo.Context) error {
	ctx := e.Request().Context()

	paymentId := e.Param("paymentId")
	err := u.app.user.DeleteCardPayment(ctx, paymentId)
	if err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}
	return e.String(http.StatusOK, paymentId)
}

func (u userServer) UpdateDefaultPayment(e echo.Context) error {
	return nil
}
