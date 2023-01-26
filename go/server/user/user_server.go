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
	GetUserPaymentPoint(ctx context.Context, userId string) (entity.UserPaymentPoint, error)

	ListTags(context.Context) ([]value.Tag, error)
	ListTaxiCallRequest(context.Context, request.ListUserTaxiCallRequest) ([]entity.TaxiCallRequest, string, error)
	GetLatestTaxiCallRequest(context.Context, string) (entity.UserLatestTaxiCallRequest, error)
	CreateTaxiCallRequest(context.Context, request.CreateTaxiCallRequest) (entity.TaxiCallRequest, error)
	CancelTaxiCallRequest(context.Context, request.UserCancelTaxiCallRequest) error
	GetUserLatestTaxiCallTicket(ctx context.Context, taxiCallRequestId string) (entity.TaxiCallTicket, error)

	SearchLocation(context.Context, request.SearchLocationRequest) ([]value.LocationSummary, int, error)
	GetAddress(context.Context, request.GetAddressRequest) (value.Address, error)

	ListAvailableServiceRegion(context.Context) ([]string, error)
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
	if err != nil {
		return server.ToResponse(e, err)
	}

	resp := response.UserSignupResponse{
		Token: token,
		User:  response.UserToResponse(user),
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

	resp := response.UserToResponse(user)

	return e.JSON(http.StatusOK, resp)
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

	resp := response.UserToResponse(user)

	return e.JSON(http.StatusOK, resp)
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

	resp := response.UserTaxiCallRequestToResponse(taxiCallRequest)
	return e.JSON(http.StatusOK, resp)
}

func (u userServer) GetLatestTaxiCallRequest(e echo.Context) error {
	ctx := e.Request().Context()

	userId := e.Param("userId")

	taxiCallRequest, err := u.app.user.GetLatestTaxiCallRequest(ctx, userId)
	if err != nil {
		return server.ToResponse(e, err)
	}

	resp := response.UserLatestTaxiCallRequestToResponse(taxiCallRequest)
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

	resp := slices.Map(taxiCallRequests, response.UserTaxiCallRequestToHistoryResponse)

	return e.JSON(http.StatusOK, response.UserTaxiCallRequestHistoryPageResponse{
		PageToken: pageToken,
		Data:      resp,
	})
}

func (u userServer) CancelTaxiCallRequest(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.UserCancelTaxiCallRequest{}

	if err := e.Bind(&req); err != nil {
		return err
	}

	if err := u.app.user.CancelTaxiCallRequest(ctx, req); err != nil {
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

	locations, pageToken, err := u.app.user.SearchLocation(ctx, req)
	if err != nil {
		return server.ToResponse(e, err)
	}

	resp := response.SearchLocationsToResopnse(locations, pageToken)

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

func (u userServer) GetUserPaymentPoint(e echo.Context) error {
	ctx := e.Request().Context()

	userId := e.Param("userId")

	userPoint, err := u.app.user.GetUserPaymentPoint(ctx, userId)
	if err != nil {
		return server.ToResponse(e, err)
	}

	resp := response.UserPaymentPointToResponse(userPoint)

	return e.JSON(http.StatusOK, resp)
}

func (u userServer) ListAvailableServiceRegion(e echo.Context) error {
	ctx := e.Request().Context()

	resp, err := u.app.user.ListAvailableServiceRegion(ctx)
	if err != nil {
		return server.ToResponse(e, err)
	}

	return e.JSON(http.StatusOK, resp)
}

func (u userServer) GetUserLatestTaxiCallTicket(e echo.Context) error {
	ctx := e.Request().Context()

	taxiCallRequestId := e.Param("taxiCallRequestId")
	ticket, err := u.app.user.GetUserLatestTaxiCallTicket(ctx, taxiCallRequestId)
	if err != nil {
		return server.ToResponse(e, err)
	}

	return e.JSON(http.StatusOK, response.TaxiCallTicketToResponse(ticket))
}
