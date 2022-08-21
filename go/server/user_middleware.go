package server

import (
	"context"

	"github.com/ktk1012/taco/go/domain/entity"
	"github.com/ktk1012/taco/go/utils"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type userSessionApp interface {
	GetSession(context.Context, string) (entity.UserSession, error)
}

type sessionMiddleware struct {
	sessionApp userSessionApp
}

func (s sessionMiddleware) Get() echo.MiddlewareFunc {
	return middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		Skipper:   s.skipper,
		Validator: s.validateSession,
	})
}

func (s sessionMiddleware) skipper(c echo.Context) bool {
	return c.Path() != "/user/signup"
}

func (s sessionMiddleware) validateSession(key string, c echo.Context) (bool, error) {
	ctx := c.Request().Context()
	session, err := s.sessionApp.GetSession(ctx, key)
	if err != nil {
		// TODO(taekyeom) handle error
		return false, err
	}

	currentTime := utils.GetRequestTimeOrNow(ctx)

	if session.ExpireTime.After(currentTime) {
		return false, nil
	}

	// Set userId key
	utils.SetUserId(ctx, session.UserId)
	r := c.Request().WithContext(ctx)
	c.SetRequest(r)

	return true, nil
}
