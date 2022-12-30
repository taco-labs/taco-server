package driver

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/response"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/server"
	"github.com/taco-labs/taco/go/utils/slices"
)

type driverApp interface {
	SmsVerificationRequest(context.Context, request.SmsVerificationRequest) (entity.SmsVerification, error)
	SmsSignin(context.Context, request.SmsSigninRequest) (entity.Driver, string, error)
	Signup(context.Context, request.DriverSignupRequest) (entity.Driver, string, error)
	GetDriver(context.Context, string) (entity.Driver, error)
	GetDriverImageUrls(context.Context, string) (value.DriverImageUrls, value.DriverImageUrls, error)
	UpdateDriver(context.Context, request.DriverUpdateRequest) (entity.Driver, error)
	UpdateOnDuty(context.Context, request.DriverOnDutyUpdateRequest) error
	UpdateDriverLocation(context.Context, request.DriverLocationUpdateRequest) error
	GetDriverSettlementAccount(context.Context, string) (entity.DriverSettlementAccount, error)
	RegisterDriverSettlementAccount(context.Context,
		request.DriverSettlementAccountRegisterRequest) (entity.DriverSettlementAccount, error)
	UpdateDriverSettlementAccount(context.Context,
		request.DriverSettlementAccountUpdateRequest) (entity.DriverSettlementAccount, error)
	ActivateDriver(context.Context, string) error
	ListTaxiCallRequest(context.Context, request.ListDriverTaxiCallRequest) ([]entity.TaxiCallRequest, string, error)
	GetLatestTaxiCallRequest(context.Context, string) (entity.DriverLatestTaxiCallRequest, error)
	AcceptTaxiCallRequest(context.Context, string) (entity.DriverLatestTaxiCallRequest, error)
	RejectTaxiCallRequest(context.Context, string) error
	CancelTaxiCallRequest(context.Context, request.CancelTaxiCallRequest) error
	DriverToArrival(context.Context, string) error
	DoneTaxiCallRequest(context.Context, request.DoneTaxiCallRequest) error
	DriverLatestTaxiCallTicket(ctx context.Context, driverId string) (entity.DriverLatestTaxiCallRequestTicket, error)

	GetExpectedDriverSettlement(context.Context, string) (entity.DriverTotalSettlement, error)
	ListDriverSettlementHistory(context.Context, request.ListDriverSettlementHistoryRequest) ([]entity.DriverSettlementHistory, time.Time, error)
	RequestSettlementTransfer(ctx context.Context, driverId string) (int, error)

	AvailableServiceRegions(ctx context.Context) ([]string, error)

	AddDriverDenyTaxiCallTag(ctx context.Context, driverId string, tagId int) error
	DeleteDriverDenyTaxiCallTag(ctx context.Context, driverId string, tagId int) error
	ListDriverDenyTaxiCallTag(ctx context.Context, driverId string) ([]value.Tag, error)
}

func (d driverServer) SmsVerificationRequest(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.SmsVerificationRequest{}

	if err := e.Bind(&req); err != nil {
		return err
	}

	smsVerification, err := d.app.driver.SmsVerificationRequest(ctx, req)
	if err != nil {
		return server.ToResponse(e, err)
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
		return server.ToResponse(e, err)
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
		return server.ToResponse(e, err)
	}

	resp := response.DriverSignupResponse{
		Token:  token,
		Driver: response.DriverToResponse(driver),
	}

	return e.JSON(http.StatusOK, resp)
}

func (d driverServer) GetDriver(e echo.Context) error {
	ctx := e.Request().Context()

	driverId := e.Param("driverId")

	driver, err := d.app.driver.GetDriver(ctx, driverId)
	if err != nil {
		return server.ToResponse(e, err)
	}

	resp := response.DriverToResponse(driver)

	return e.JSON(http.StatusOK, resp)
}

func (d driverServer) UpdateDriver(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.DriverUpdateRequest{}

	if err := e.Bind(&req); err != nil {
		return err
	}

	driver, err := d.app.driver.UpdateDriver(ctx, req)
	if err != nil {
		return server.ToResponse(e, err)
	}

	resp := response.DriverToResponse(driver)

	return e.JSON(http.StatusOK, resp)
}

func (d driverServer) GetDriverImageUrls(e echo.Context) error {
	ctx := e.Request().Context()

	driverId := e.Param("driverId")

	downloadUrls, uploadUrls, err := d.app.driver.GetDriverImageUrls(ctx, driverId)
	if err != nil {
		return server.ToResponse(e, err)
	}

	resp := response.DriverImageUrlResponse{
		DownloadUrls: downloadUrls,
		UploadUrls:   uploadUrls,
	}

	return e.JSON(http.StatusOK, resp)
}

func (d driverServer) UpdateOnDuty(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.DriverOnDutyUpdateRequest{}

	if err := e.Bind(&req); err != nil {
		return err
	}

	err := d.app.driver.UpdateOnDuty(ctx, req)
	if err != nil {
		return server.ToResponse(e, err)
	}

	return e.JSON(http.StatusOK, struct{}{})
}

func (d driverServer) UpdateDriverLocation(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.DriverLocationUpdateRequest{}

	if err := e.Bind(&req); err != nil {
		return err
	}

	err := d.app.driver.UpdateDriverLocation(ctx, req)
	if err != nil {
		return server.ToResponse(e, err)
	}

	return e.JSON(http.StatusOK, struct{}{})
}

func (d driverServer) GetDriverSettlemtnAccount(e echo.Context) error {
	ctx := e.Request().Context()

	driverId := e.Param("driverId")

	account, err := d.app.driver.GetDriverSettlementAccount(ctx, driverId)
	if err != nil {
		return server.ToResponse(e, err)
	}

	resp := response.DriverSettlemtnAccountToResponse(account)

	return e.JSON(http.StatusOK, resp)
}

func (d driverServer) RegisterDriverSettlementAccount(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.DriverSettlementAccountRegisterRequest{}

	if err := e.Bind(&req); err != nil {
		return err
	}

	account, err := d.app.driver.RegisterDriverSettlementAccount(ctx, req)
	if err != nil {
		return server.ToResponse(e, err)
	}

	resp := response.DriverSettlemtnAccountToResponse(account)

	return e.JSON(http.StatusOK, resp)
}

func (d driverServer) UpdateDriverSettlemtnAccount(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.DriverSettlementAccountUpdateRequest{}

	if err := e.Bind(&req); err != nil {
		return err
	}

	account, err := d.app.driver.UpdateDriverSettlementAccount(ctx, req)
	if err != nil {
		return server.ToResponse(e, err)
	}

	resp := response.DriverSettlemtnAccountToResponse(account)

	return e.JSON(http.StatusOK, resp)
}

func (d driverServer) ActivateDriver(e echo.Context) error {
	ctx := e.Request().Context()

	driverId := e.Param("driverId")

	if err := d.app.driver.ActivateDriver(ctx, driverId); err != nil {
		return server.ToResponse(e, err)
	}

	return nil
}

func (d driverServer) GetLatestTaxiCallRequest(e echo.Context) error {
	ctx := e.Request().Context()

	userId := e.Param("driverId")

	taxiCallRequest, err := d.app.driver.GetLatestTaxiCallRequest(ctx, userId)
	if err != nil {
		return server.ToResponse(e, err)
	}

	resp := response.DriverLatestTaxiCallRequestToResponse(taxiCallRequest)
	return e.JSON(http.StatusOK, resp)
}

func (d driverServer) ListTaxiCallRequest(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.ListDriverTaxiCallRequest{}

	if err := e.Bind(&req); err != nil {
		return err
	}

	if req.Count == 0 {
		req.Count = 30
	}

	taxiCallRequests, pageToken, err := d.app.driver.ListTaxiCallRequest(ctx, req)
	if err != nil {
		return server.ToResponse(e, err)
	}

	resp := slices.Map(taxiCallRequests, response.DriverTaxiCallRequestToResponse)

	return e.JSON(http.StatusOK, response.DriverTaxiCallRequestPageResponse{
		PageToken: pageToken,
		Data:      resp,
	})
}

func (d driverServer) AcceptTaxiCallRequest(e echo.Context) error {
	ctx := e.Request().Context()

	ticketId := e.Param("ticketId")

	taxiCallRequest, err := d.app.driver.AcceptTaxiCallRequest(ctx, ticketId)
	if err != nil {
		return server.ToResponse(e, err)
	}

	resp := response.DriverLatestTaxiCallRequestToResponse(taxiCallRequest)

	return e.JSON(http.StatusOK, resp)
}

func (d driverServer) RejectTaxiCallRequest(e echo.Context) error {
	ctx := e.Request().Context()

	ticketId := e.Param("ticketId")

	err := d.app.driver.RejectTaxiCallRequest(ctx, ticketId)
	if err != nil {
		return server.ToResponse(e, err)
	}

	return e.JSON(http.StatusOK, struct{}{})
}

func (d driverServer) CancelTaxiCallRequest(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.CancelTaxiCallRequest{}

	if err := e.Bind(&req); err != nil {
		return err
	}

	err := d.app.driver.CancelTaxiCallRequest(ctx, req)
	if err != nil {
		return server.ToResponse(e, err)
	}

	return e.JSON(http.StatusOK, struct{}{})
}

func (d driverServer) DriverToArrival(e echo.Context) error {
	ctx := e.Request().Context()

	taxiCallRequestId := e.Param("taxiCallRequestId")

	err := d.app.driver.DriverToArrival(ctx, taxiCallRequestId)
	if err != nil {
		return server.ToResponse(e, err)
	}

	return e.JSON(http.StatusOK, struct{}{})
}

func (d driverServer) DoneTaxiCallRequest(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.DoneTaxiCallRequest{}

	if err := e.Bind(&req); err != nil {
		return err
	}

	err := d.app.driver.DoneTaxiCallRequest(ctx, req)
	if err != nil {
		return server.ToResponse(e, err)
	}

	return e.JSON(http.StatusOK, struct{}{})
}

func (d driverServer) LatestTaxiCallTicket(e echo.Context) error {
	ctx := e.Request().Context()

	driverId := e.Param("driverId")

	latestTicket, err := d.app.driver.DriverLatestTaxiCallTicket(ctx, driverId)
	if err != nil {
		return server.ToResponse(e, err)
	}

	resp := response.DriverLatestTaxiCallRequestTicketToResponse(latestTicket)

	return e.JSON(http.StatusOK, resp)
}

func (d driverServer) GetExpectedDriverSetttlement(e echo.Context) error {
	ctx := e.Request().Context()

	driverId := e.Param("driverId")

	resp, err := d.app.driver.GetExpectedDriverSettlement(ctx, driverId)
	if err != nil {
		return server.ToResponse(e, err)
	}

	return e.JSON(http.StatusOK, response.DriverTotalSettlementToResponse(resp))
}

func (d driverServer) ListDriverSettlementHistory(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.ListDriverSettlementHistoryRequest{}

	if err := e.Bind(&req); err != nil {
		return err
	}

	histories, pageToken, err := d.app.driver.ListDriverSettlementHistory(ctx, req)
	if err != nil {
		return server.ToResponse(e, err)
	}

	resp := response.ListDriverSettlementHistoryToResponse(histories, pageToken)

	return e.JSON(http.StatusOK, resp)
}

func (d driverServer) RequestDriverSettlementTransfer(e echo.Context) error {
	ctx := e.Request().Context()

	driverId := e.Param("driverId")

	resp, err := d.app.driver.RequestSettlementTransfer(ctx, driverId)
	if err != nil {
		return server.ToResponse(e, err)
	}

	return e.JSON(http.StatusOK, response.DriverSettlementTransferResponse{
		ExpectedTransferAmount: resp,
	})
}

func (d driverServer) AvailableServiceRegions(e echo.Context) error {
	ctx := e.Request().Context()

	resp, err := d.app.driver.AvailableServiceRegions(ctx)
	if err != nil {
		return server.ToResponse(e, err)
	}

	return e.JSON(http.StatusOK, resp)
}

func (d driverServer) AddDriverDenyTaxiCallTag(e echo.Context) error {
	ctx := e.Request().Context()

	driverId := e.Param("driverId")
	tagId, err := strconv.Atoi(e.Param("tagId"))
	if err != nil {
		tagId = -1
	}

	err = d.app.driver.AddDriverDenyTaxiCallTag(ctx, driverId, tagId)
	if err != nil {
		return server.ToResponse(e, err)
	}

	return e.JSON(http.StatusOK, struct{}{})
}

func (d driverServer) DeleteDriverDenyTaxiCallTag(e echo.Context) error {
	ctx := e.Request().Context()

	driverId := e.Param("driverId")
	tagId, err := strconv.Atoi(e.Param("tagId"))
	if err != nil {
		tagId = -1
	}

	err = d.app.driver.DeleteDriverDenyTaxiCallTag(ctx, driverId, tagId)
	if err != nil {
		return server.ToResponse(e, err)
	}

	return e.JSON(http.StatusOK, struct{}{})
}

func (d driverServer) ListDriverDenyTaxiCallTag(e echo.Context) error {
	ctx := e.Request().Context()

	driverId := e.Param("driverId")

	resp, err := d.app.driver.ListDriverDenyTaxiCallTag(ctx, driverId)
	if err != nil {
		return server.ToResponse(e, err)
	}

	return e.JSON(http.StatusOK, resp)
}
