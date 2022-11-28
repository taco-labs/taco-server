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
	UpdateUser(context.Context, request.UserUpdateRequest) (entity.User, error)

	ListUserPayment(context.Context, string) ([]entity.UserPayment, error)
	TryRecoverUserPayment(context.Context, string) error
	DeleteUserPayment(context.Context, string) error

	ListTags(context.Context) ([]value.Tag, error)
	ListTaxiCallRequest(context.Context, request.ListUserTaxiCallRequest) ([]entity.TaxiCallRequest, string, error)
	GetLatestTaxiCallRequest(context.Context, string) (entity.TaxiCallRequest, error)
	CreateTaxiCallRequest(context.Context, request.CreateTaxiCallRequest) (entity.TaxiCallRequest, error)
	CancelTaxiCallRequest(context.Context, string) error

	SearchLocation(context.Context, request.SearchLocationRequest) ([]value.LocationSummary, error)
	GetAddress(context.Context, request.GetAddressRequest) (value.Address, error)
}

func (u userServer) SmsVerificationRequest(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.SmsVerificationRequest{}

	if err := e.Bind(&req); err != nil {
		return err
	}

	smsVerification, err := u.app.user.SmsVerificationRequest(ctx, req)
	if err != nil {
		return server.ToResponse(e, err)
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
		return server.ToResponse(e, err)
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
		return server.ToResponse(e, err)
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
		return server.ToResponse(e, err)
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
		return server.ToResponse(e, err)
	}

	return e.JSON(http.StatusOK, response.UserToResponse(user))
}

func (u userServer) ListUserPayment(e echo.Context) error {
	ctx := e.Request().Context()
	userId := e.Param("userId")
	userPayments, err := u.app.user.ListUserPayment(ctx, userId)
	if err != nil {
		return server.ToResponse(e, err)
	}

	return e.JSON(http.StatusOK, response.ListUserPaymentResponse{
		Payments: slices.Map(userPayments, response.UserPaymentToResponse),
	})
}

func (u userServer) TryRecoverUserPayment(e echo.Context) error {
	ctx := e.Request().Context()

	paymentId := e.Param("paymentId")
	err := u.app.user.TryRecoverUserPayment(ctx, paymentId)
	if err != nil {
		return server.ToResponse(e, err)
	}

	return e.JSON(http.StatusOK, struct{}{})
}

func (u userServer) DeleteUserPayment(e echo.Context) error {
	ctx := e.Request().Context()

	paymentId := e.Param("paymentId")
	err := u.app.user.DeleteUserPayment(ctx, paymentId)
	if err != nil {
		return server.ToResponse(e, err)
	}
	return e.JSON(http.StatusOK, response.DeleteUserPaymentResponse{
		PaymentId: paymentId,
	})
}

func (u userServer) CreateTaxiCallRequest(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.CreateTaxiCallRequest{}
	if err := e.Bind(&req); err != nil {
		return err
	}

	taxiCallRequest, err := u.app.user.CreateTaxiCallRequest(ctx, req)
	if err != nil {
		return server.ToResponse(e, err)
	}

	resp := response.TaxiCallRequestToResponse(taxiCallRequest)
	return e.JSON(http.StatusOK, resp)
}

func (u userServer) GetLatestTaxiCallRequest(e echo.Context) error {
	ctx := e.Request().Context()

	userId := e.Param("userId")

	taxiCallRequest, err := u.app.user.GetLatestTaxiCallRequest(ctx, userId)
	if err != nil {
		return server.ToResponse(e, err)
	}

	resp := response.TaxiCallRequestToResponse(taxiCallRequest)
	return e.JSON(http.StatusOK, resp)
}

func (u userServer) ListTags(e echo.Context) error {
	ctx := e.Request().Context()

	tags, err := u.app.user.ListTags(ctx)
	if err != nil {
		return server.ToResponse(e, err)
	}

	return e.JSON(http.StatusOK, tags)
}

func (u userServer) ListTaxiCallRequest(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.ListUserTaxiCallRequest{}

	if err := e.Bind(&req); err != nil {
		return err
	}

	if req.Count == 0 {
		req.Count = 30
	}

	taxiCallRequests, pageToken, err := u.app.user.ListTaxiCallRequest(ctx, req)
	if err != nil {
		return server.ToResponse(e, err)
	}

	resp := slices.Map(taxiCallRequests, response.TaxiCallRequestToResponse)

	return e.JSON(http.StatusOK, response.TaxiCallRequestPageResponse{
		PageToken: pageToken,
		Data:      resp,
	})
}

func (u userServer) CancelTaxiCallRequest(e echo.Context) error {
	ctx := e.Request().Context()

	taxiCallRequestId := e.Param("taxiCallRequestId")

	if err := u.app.user.CancelTaxiCallRequest(ctx, taxiCallRequestId); err != nil {
		return server.ToResponse(e, err)
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
		return server.ToResponse(e, err)
	}

	return e.JSON(http.StatusOK, resp)
}

func (u userServer) GetAddress(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.GetAddressRequest{}
	if err := e.Bind(&req); err != nil {
		return err
	}

	resp, err := u.app.user.GetAddress(ctx, req)
	if err != nil {
		return server.ToResponse(e, err)
	}

	return e.JSON(http.StatusOK, resp)
}
