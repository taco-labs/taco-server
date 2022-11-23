package driver

import (
	"context"
	"errors"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/server"
	"github.com/taco-labs/taco/go/utils"
)

var skipSet = map[string]struct{}{
	"/driver/signin/sms/request": {},
	"/driver/signin/sms/verify":  {},
	"/driver/signup":             {},
	"/healthz":                   {},
}

var denyNonActiveDriverPath = map[string]struct{}{
	"/driver/:driverId/on_duty":               {},
	"/driver/:driverId/location":              {},
	"/taxicall/ticket/:ticketId":              {},
	"/taxicall/:taxiCallRequestId":            {},
	"/taxicall/:taxiCallRequestId/to_arrival": {},
	"/taxicall/:taxiCallRequestId/done":       {},
}

type driverSessionApp interface {
	GetById(context.Context, string) (entity.DriverSession, error)
}

type sessionMiddleware struct {
	sessionApp driverSessionApp
}

func (s sessionMiddleware) Get() echo.MiddlewareFunc {
	return middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		Skipper:   s.skipper,
		Validator: s.validateSession,
		ErrorHandler: func(err error, c echo.Context) error {
			return server.ToResponse(err)
		},
	})
}

func (s sessionMiddleware) skipper(c echo.Context) bool {
	_, ok := skipSet[c.Path()]
	return ok
}

func (s sessionMiddleware) validateSession(key string, c echo.Context) (bool, error) {
	ctx := c.Request().Context()
	session, err := s.sessionApp.GetById(ctx, key)
	if errors.Is(err, value.ErrNotFound) {
		return false, value.ErrUnAuthenticated
	}
	if err != nil {
		return false, err
	}

	_, denyPath := denyNonActiveDriverPath[c.Path()]
	if !session.Activated && denyPath {
		return false, value.ErrNotYetActivated
	}

	currentTime := utils.GetRequestTimeOrNow(ctx)

	if session.ExpireTime.Before(currentTime) {
		return false, value.ErrSessionExpired
	}

	ctx = utils.SetDriverId(ctx, session.DriverId)
	r := c.Request().WithContext(ctx)
	c.SetRequest(r)

	return true, nil
}

func NewSessionMiddleware(sessionApp driverSessionApp) sessionMiddleware {
	return sessionMiddleware{
		sessionApp: sessionApp,
	}
}

func DriverIdChecker(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		requestDriverId := c.Param("driverId")

		if requestDriverId != "" {
			driverId := utils.GetDriverId(ctx)
			if requestDriverId != driverId {
				return server.ToResponse(fmt.Errorf("unauthorized access to driver resource:%w", value.ErrUnAuthorized))
			}
		}

		return next(c)
	}
}
