package user

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/response"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/server"
	"github.com/taco-labs/taco/go/utils/slices"
)

type UserApp interface {
	SmsVerificationRequest(context.Context, request.SmsVerificationRequest) (entity.SmsVerification, error)
	SmsSignin(context.Context, request.SmsSigninRequest) (entity.User, string, error)
	Signup(context.Context, request.UserSignupRequest) (entity.User, string, error)
	GetUser(context.Context, string) (entity.User, error)
	DeleteUser(context.Context, string) error
	UpdateUser(context.Context, request.UserUpdateRequest) (entity.User, error)
	ListCardPayment(ctx context.Context, userId string) ([]entity.UserPayment, entity.UserDefaultPayment, error)
	RegisterCardPayment(ctx context.Context, req request.UserPaymentRegisterRequest) (entity.UserPayment, error)
	DeleteCardPayment(ctx context.Context, userPaymentId string) error
	UpdateDefaultPayment(ctx context.Context, req request.DefaultPaymentUpdateRequest) error
	ListTaxiCallRequest(context.Context, string) ([]entity.TaxiCallRequest, error)
	GetLatestTaxiCallRequest(context.Context, string) (entity.TaxiCallRequest, error)
	CreateTaxiCallRequest(context.Context, request.CreateTaxiCallRequest) (entity.TaxiCallRequest, error)
	CancelTaxiCallRequest(context.Context, string) error
	SearchLocation(context.Context, request.SearchLocationRequest) ([]value.LocationSummary, error)
}

func (u userServer) SmsVerificationRequest(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.SmsVerificationRequest{}

	if err := e.Bind(&req); err != nil {
		return err
	}

	smsVerification, err := u.app.user.SmsVerificationRequest(ctx, req)
	if err != nil {
		return server.ToResponse(err)
	}

	resp := response.SmsVerificationRequestResponse{
		Id:         smsVerification.Id,
		ExpireTime: smsVerification.ExpireTime,
	}

	return e.JSON(http.StatusOK, resp)
}

func (u userServer) SmsSignin(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.SmsSigninRequest{}

	if err := e.Bind(&req); err != nil {
		return err
	}

	user, token, err := u.app.user.SmsSignin(ctx, req)

	resp := response.UserSignupResponse{
		Token: token,
		User:  response.UserToResponse(user),
	}

	if err != nil {
		return server.ToResponse(err)
	}

	return e.JSON(http.StatusOK, resp)
}

func (u userServer) Signup(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.UserSignupRequest{}

	if err := e.Bind(&req); err != nil {
		// TODO(taekyeom) Error handle
		return err
	}

	user, token, err := u.app.user.Signup(ctx, req)

	if err != nil {
		return server.ToResponse(err)
	}

	resp := response.UserSignupResponse{
		Token: token,
		User:  response.UserToResponse(user),
	}

	return e.JSON(http.StatusOK, resp)
}

func (u userServer) GetUser(e echo.Context) error {
	ctx := e.Request().Context()
	userId := e.Param("userId")
	user, err := u.app.user.GetUser(ctx, userId)
	if err != nil {
		return server.ToResponse(err)
	}
	return e.JSON(http.StatusOK, response.UserToResponse(user))
}

func (u userServer) UpdateUser(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.UserUpdateRequest{}

	if err := e.Bind(&req); err != nil {
		return err
	}

	user, err := u.app.user.UpdateUser(ctx, req)
	if err != nil {
		return server.ToResponse(err)
	}

	return e.JSON(http.StatusOK, response.UserToResponse(user))
}

func (u userServer) DeleteUser(e echo.Context) error {
	ctx := e.Request().Context()

	userId := e.Param("userId")
	err := u.app.user.DeleteUser(ctx, userId)
	if err != nil {
		return server.ToResponse(err)
	}

	return e.JSON(http.StatusOK, struct{}{})
}

func (u userServer) ListCardPayment(e echo.Context) error {
	ctx := e.Request().Context()
	userId := e.Param("userId")
	userPayments, userDefaultPayment, err := u.app.user.ListCardPayment(ctx, userId)
	if err != nil {
		return server.ToResponse(err)
	}

	return e.JSON(http.StatusOK, response.ListCardPaymentResponse{
		DefaultPaymentId: userDefaultPayment.PaymentId,
		Payments:         slices.Map(userPayments, response.UserPaymentToResponse),
	})
}

func (u userServer) RegisterCardPayment(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.UserPaymentRegisterRequest{}
	if err := e.Bind(&req); err != nil {
		return err
	}

	cardPayment, err := u.app.user.RegisterCardPayment(ctx, req)
	if err != nil {
		return server.ToResponse(err)
	}

	return e.JSON(http.StatusOK, response.UserPaymentToResponse(cardPayment))
}

func (u userServer) DeleteCardPayment(e echo.Context) error {
	ctx := e.Request().Context()

	paymentId := e.Param("paymentId")
	err := u.app.user.DeleteCardPayment(ctx, paymentId)
	if err != nil {
		return server.ToResponse(err)
	}
	return e.JSON(http.StatusOK, response.DeleteCardPaymentResponse{
		PaymentId: paymentId,
	})
}

func (u userServer) UpdateDefaultPayment(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.DefaultPaymentUpdateRequest{}

	if err := e.Bind(&req); err != nil {
		return err
	}

	if err := u.app.user.UpdateDefaultPayment(ctx, req); err != nil {
		return server.ToResponse(err)
	}

	return e.JSON(http.StatusOK, struct{}{})
}

func (u userServer) CreateTaxiCallRequest(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.CreateTaxiCallRequest{}
	if err := e.Bind(&req); err != nil {
		return err
	}

	taxiCallRequest, err := u.app.user.CreateTaxiCallRequest(ctx, req)
	if err != nil {
		return server.ToResponse(err)
	}

	resp := response.TaxiCallRequestToResponse(taxiCallRequest)
	return e.JSON(http.StatusOK, resp)
}

func (u userServer) GetLatestTaxiCallRequest(e echo.Context) error {
	ctx := e.Request().Context()

	userId := e.Param("userId")

	taxiCallRequest, err := u.app.user.GetLatestTaxiCallRequest(ctx, userId)
	if err != nil {
		return server.ToResponse(err)
	}

	resp := response.TaxiCallRequestToResponse(taxiCallRequest)
	return e.JSON(http.StatusOK, resp)
}

func (u userServer) ListTaxiCallRequest(e echo.Context) error {
	ctx := e.Request().Context()

	userId := e.Param("userId")

	taxiCallRequests, err := u.app.user.ListTaxiCallRequest(ctx, userId)
	if err != nil {
		return server.ToResponse(err)
	}

	resp := slices.Map(taxiCallRequests, response.TaxiCallRequestToResponse)

	return e.JSON(http.StatusOK, response.TaxiCallRequestPageResponse{
		Data: resp,
	})
}

func (u userServer) CancelTaxiCallRequest(e echo.Context) error {
	ctx := e.Request().Context()

	taxiCallRequestId := e.Param("taxiCallRequestId")

	if err := u.app.user.CancelTaxiCallRequest(ctx, taxiCallRequestId); err != nil {
		return server.ToResponse(err)
	}

	return e.JSON(http.StatusOK, struct{}{})
}

func (u userServer) SearchLocation(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.SearchLocationRequest{}
	if err := e.Bind(&req); err != nil {
		return err
	}

	resp, err := u.app.user.SearchLocation(ctx, req)
	if err != nil {
		return server.ToResponse(err)
	}

	return e.JSON(http.StatusOK, resp)
}
