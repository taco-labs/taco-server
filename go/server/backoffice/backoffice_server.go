package backoffice

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/response"
	"github.com/taco-labs/taco/go/server"
	"github.com/taco-labs/taco/go/utils"
	"github.com/taco-labs/taco/go/utils/slices"
)

func (b backofficeServer) GetDriver(e echo.Context) error {
	ctx := e.Request().Context()

	driverId := e.Param("driverId")

	driver, err := b.app.driver.GetDriver(ctx, driverId)
	if err != nil {
		return server.ToResponse(e, err)
	}

	resp := response.DriverToResponse(driver)

	return e.JSON(http.StatusOK, resp)
}

func (b backofficeServer) DeleteDriver(e echo.Context) error {
	ctx := e.Request().Context()

	driverId := e.Param("driverId")

	err := b.app.driver.DeleteDriver(ctx, driverId)
	if err != nil {
		return server.ToResponse(e, err)
	}

	return e.JSON(http.StatusOK, struct{}{})
}

func (b backofficeServer) ActivateDriver(e echo.Context) error {
	ctx := e.Request().Context()

	driverId := e.Param("driverId")

	err := b.app.driver.ActivateDriver(ctx, driverId)
	if err != nil {
		return server.ToResponse(e, err)
	}

	return e.JSON(http.StatusOK, struct{}{})
}

func (b backofficeServer) GetUser(e echo.Context) error {
	ctx := e.Request().Context()
	userId := e.Param("userId")
	user, err := b.app.user.GetUser(ctx, userId)
	if err != nil {
		return server.ToResponse(e, err)
	}

	resp := response.UserToResponse(user)

	return e.JSON(http.StatusOK, resp)
}

func (b backofficeServer) DeleteUser(e echo.Context) error {
	ctx := e.Request().Context()

	userId := e.Param("userId")
	err := b.app.user.DeleteUser(ctx, userId)
	if err != nil {
		return server.ToResponse(e, err)
	}

	return e.JSON(http.StatusOK, struct{}{})
}

// TODO (taekyeom) Must remove before production
func (b backofficeServer) ForceAcceptTaxiCallRequest(e echo.Context) error {
	ctx := e.Request().Context()

	driverId := e.Param("driverId")
	taxiCallRequestId := e.Param("taxiCallRequestId")

	driverLatestTaxiCallRequest, err := b.app.driver.ForceAcceptTaxiCallRequest(ctx, driverId, taxiCallRequestId)
	if err != nil {
		return server.ToResponse(e, err)
	}

	resp := response.DriverLatestTaxiCallRequestToResponse(driverLatestTaxiCallRequest)

	return e.JSON(http.StatusOK, resp)
}

func (b backofficeServer) DriverToArrival(e echo.Context) error {
	ctx := e.Request().Context()

	driverId := e.Param("driverId")
	taxiCallRequestId := e.Param("taxiCallRequestId")

	ctx = utils.SetDriverId(ctx, driverId)

	err := b.app.driver.DriverToArrival(ctx, taxiCallRequestId)
	if err != nil {
		return server.ToResponse(e, err)
	}

	return e.JSON(http.StatusOK, struct{}{})
}

func (b backofficeServer) DoneTaxiCallRequest(e echo.Context) error {
	ctx := e.Request().Context()

	driverId := e.Param("driverId")

	req := request.DoneTaxiCallRequest{}
	if err := e.Bind(&req); err != nil {
		return err
	}

	ctx = utils.SetDriverId(ctx, driverId)

	err := b.app.driver.DoneTaxiCallRequest(ctx, req)
	if err != nil {
		return server.ToResponse(e, err)
	}

	return e.JSON(http.StatusOK, struct{}{})
}

func (b backofficeServer) ListNonActivatedDriver(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.ListNonActivatedDriverRequest{}
	if err := e.Bind(&req); err != nil {
		return err
	}

	driverDtos, pageToken, err := b.app.driver.ListNonActivatedDriver(ctx, req)
	if err != nil {
		return server.ToResponse(e, err)
	}

	resp := response.ListNonActivatedDriverResponse{
		PageToken: pageToken,
		Drivers:   slices.Map(driverDtos, response.DriverDtoToResponse),
	}

	return e.JSON(http.StatusOK, resp)
}

func (b backofficeServer) GetDriverSettlementAccount(e echo.Context) error {
	ctx := e.Request().Context()

	driverId := e.Param("driverId")

	account, err := b.app.driver.GetDriverSettlementAccount(ctx, driverId)
	if err != nil {
		return server.ToResponse(e, err)
	}

	resp := response.DriverSettlemtnAccountToResponse(account)

	return e.JSON(http.StatusOK, resp)
}

func (b backofficeServer) ListDriverTaxiCallContextInRadius(e echo.Context) error {
	ctx := e.Request().Context()

	req := request.ListDriverTaxiCallContextInRadiusRequest{}

	if err := e.Bind(&req); err != nil {
		return err
	}

	driverContexts, err := b.app.taxicall.ListDriverTaxiCallContextInRadius(ctx, req)
	if err != nil {
		return server.ToResponse(e, err)
	}

	resp := slices.Map(driverContexts, response.DriverTaxiCallContextToResponse)

	return e.JSON(http.StatusOK, resp)
}
