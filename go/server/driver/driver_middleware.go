package driver

import (
	"context"
	"errors"
	"fmt"
	"time"

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
	"/service_region":            {},
	"/download":                  {},

	// TODO (taekyeom) move to extension
	"/payment/payple/settlement_callback": {},
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
	Update(context.Context, entity.DriverSession) error
}

type sessionMiddleware struct {
	sessionApp driverSessionApp
}

func (s sessionMiddleware) Get() echo.MiddlewareFunc {
	return middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		Skipper:   s.skipper,
		Validator: s.validateSession,
		ErrorHandler: func(err error, c echo.Context) error {
			tacoErr := value.TacoError{}
			if errors.As(err, &tacoErr) {
				return server.ToResponse(c, err)
			}
			wrappedErr := fmt.Errorf("%w: error from key auth: %v", value.ErrUnAuthenticated, err)
			return server.ToResponse(c, wrappedErr)
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

	// If we need to extend expiration time, update it
	if session.ExpireTime.Sub(currentTime) < (time.Hour*24)*7 {
		session.ExpireTime = currentTime.AddDate(0, 1, 0)
		if err := s.sessionApp.Update(ctx, session); err != nil {
			return false, err
		}
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
				return server.ToResponse(c, fmt.Errorf("unauthorized access to driver resource:%w", value.ErrUnAuthorized))
			}
		}

		return next(c)
	}
}
