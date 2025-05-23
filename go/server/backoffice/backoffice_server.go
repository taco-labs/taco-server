package backoffice

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/response"
	"github.com/taco-labs/taco/go/server"
	"github.com/taco-labs/taco/go/utils"
)

func (b backofficeServer) GetDriver(e echo.Context) error {
	ctx := e.Request().Context()

	driverId := e.Param("driverId")

	driver, err := b.app.driver.GetDriver(ctx, driverId)
	if err != nil {
		return server.ToResponse(err)
	}

	resp := response.DriverToResponse(driver)

	return e.JSON(http.StatusOK, resp)
}

func (b backofficeServer) DeleteDriver(e echo.Context) error {
	ctx := e.Request().Context()

	driverId := e.Param("driverId")

	err := b.app.driver.DeleteDriver(ctx, driverId)
	if err != nil {
		return server.ToResponse(err)
	}

	return e.JSON(http.StatusOK, struct{}{})
}

func (b backofficeServer) ActivateDriver(e echo.Context) error {
	ctx := e.Request().Context()

	driverId := e.Param("driverId")

	err := b.app.driver.ActivateDriver(ctx, driverId)
	if err != nil {
		return server.ToResponse(err)
	}

	return e.JSON(http.StatusOK, struct{}{})
}

func (b backofficeServer) GetUser(e echo.Context) error {
	ctx := e.Request().Context()
	userId := e.Param("userId")
	user, err := b.app.user.GetUser(ctx, userId)
	if err != nil {
		return server.ToResponse(err)
	}
	return e.JSON(http.StatusOK, response.UserToResponse(user))
}

func (b backofficeServer) DeleteUser(e echo.Context) error {
	ctx := e.Request().Context()

	userId := e.Param("userId")
	err := b.app.user.DeleteUser(ctx, userId)
	if err != nil {
		return server.ToResponse(err)
	}

	return e.JSON(http.StatusOK, struct{}{})
}

// TODO (taekyeom) Must remove before production
func (b backofficeServer) ForceAcceptTaxiCallRequest(e echo.Context) error {
	ctx := e.Request().Context()

	driverId := e.Param("driverId")
	taxiCallRequestId := e.Param("taxiCallRequestId")

	err := b.app.driver.ForceAcceptTaxiCallRequest(ctx, driverId, taxiCallRequestId)
	if err != nil {
		return server.ToResponse(err)
	}

	return e.JSON(http.StatusOK, struct{}{})
}

func (b backofficeServer) DriverToArrival(e echo.Context) error {
	ctx := e.Request().Context()

	taxiCallRequestId := e.Param("taxiCallRequestId")

	err := b.app.driver.DriverToArrival(ctx, taxiCallRequestId)
	if err != nil {
		return server.ToResponse(err)
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
		return server.ToResponse(err)
	}

	return e.JSON(http.StatusOK, struct{}{})
}
