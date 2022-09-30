package user

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/response"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/utils"
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

	if errors.Is(err, value.ErrUserNotFound) {
		return e.JSON(http.StatusNotFound, value.ErrUserNotFound)
	}
	if err != nil {
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

	if err != nil {
		// TODO(taekyeom) Error handle
		return e.String(http.StatusBadRequest, err.Error())
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

func (u userServer) DeleteUser(e echo.Context) error {
	ctx := e.Request().Context()

	userId := e.Param("userId")
	err := u.app.user.DeleteUser(ctx, userId)
	if err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}

	return nil
}

func (u userServer) ListCardPayment(e echo.Context) error {
	ctx := e.Request().Context()
	userId := e.Param("userId")
	userPayments, userDefaultPayment, err := u.app.user.ListCardPayment(ctx, userId)
	if err != nil {
		return e.String(http.StatusBadRequest, err.Error())
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
	ctx := e.Request().Context()

	req := request.DefaultPaymentUpdateRequest{}

	if err := e.Bind(&req); err != nil {
		return e.String(http.StatusBadRequest, fmt.Errorf("bind error: %w", err).Error())
	}

	if err := u.app.user.UpdateDefaultPayment(ctx, req); err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}

	return e.JSON(http.StatusOK, struct{}{})
}

func (u userServer) CreateTaxiCallRequest(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.CreateTaxiCallRequest{}
	if err := e.Bind(&req); err != nil {
		return e.String(http.StatusBadRequest, fmt.Errorf("bind error: %v", err).Error())
	}

	taxiCallRequest, err := u.app.user.CreateTaxiCallRequest(ctx, req)
	if err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}

	resp := response.TaxiCallRequestToResponse(taxiCallRequest)
	return e.JSON(http.StatusOK, resp)
}

func (u userServer) GetLatestTaxiCallRequest(e echo.Context) error {
	ctx := e.Request().Context()

	userId := e.Param("userId")

	taxiCallRequest, err := u.app.user.GetLatestTaxiCallRequest(ctx, userId)
	if errors.Is(err, value.ErrNotFound) {
		return e.JSON(http.StatusNotFound, value.ErrNotFound)
	}
	if err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}

	resp := response.TaxiCallRequestToResponse(taxiCallRequest)
	return e.JSON(http.StatusOK, resp)
}

func (u userServer) ListTaxiCallRequest(e echo.Context) error {
	ctx := e.Request().Context()

	userId := e.Param("userId")

	taxiCallRequests, err := u.app.user.ListTaxiCallRequest(ctx, userId)
	if err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}

	resp := slices.Map(taxiCallRequests, response.TaxiCallRequestToResponse)

	// TODO(taekyeom) 임시 mock data
	id := utils.MustNewUUID()
	driverId := utils.MustNewUUID()
	taxiCallRequestMock := response.TaxiCallRequestResponse{
		Id:       id,
		UserId:   userId,
		DriverId: &driverId,
		Departure: value.Location{
			Latitude:  35.97664845766847,
			Longitude: 126.99597295767953,
			RoadAddress: value.RoadAddress{
				AddressName:  "전북 익산시 망산길 11-17",
				RegionDepth1: "전북",
				RegionDepth2: "익산시",
				RegionDepth3: "부송동",
				RoadName:     "망산길",
				BuildingName: "",
			},
		},
		Arrival: value.Location{
			Latitude:  37.0789561558879,
			Longitude: 127.423084873712,
			RoadAddress: value.RoadAddress{
				AddressName:  "경기도 안성시 죽산면 죽산초교길 69-4",
				RegionDepth1: "경기",
				RegionDepth2: "안성시",
				RegionDepth3: "죽산면",
				RoadName:     "죽산초교길",
				BuildingName: "무지개아파트",
			},
		},
		RequestBasePrice:          12000,
		RequestMinAdditionalPrice: 0,
		RequestMaxAdditionalPrice: 12000,
		BasePrice:                 12300,
		AdditionalPrice:           3000,
		Payment: response.PaymentSummaryResponse{
			PaymentId:  utils.MustNewUUID(),
			Company:    "현대",
			CardNumber: "433012******1234",
		},
	}

	return e.JSON(http.StatusOK, response.TaxiCallRequestPageResponse{
		Data: append(resp, taxiCallRequestMock),
	})
}

func (u userServer) CancelTaxiCallRequest(e echo.Context) error {
	ctx := e.Request().Context()

	taxiCallRequestId := e.Param("taxiCallRequestId")

	if err := u.app.user.CancelTaxiCallRequest(ctx, taxiCallRequestId); err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}

	return e.JSON(http.StatusOK, struct{}{})
}
