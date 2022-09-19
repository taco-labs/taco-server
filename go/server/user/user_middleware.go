package user

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/utils"
)

var skipSet = map[string]struct{}{
	"/user/signup": {},
	"/healthz":     {},
}

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
	_, ok := skipSet[c.Path()]
	return ok
}

func (s sessionMiddleware) validateSession(key string, c echo.Context) (bool, error) {
	ctx := c.Request().Context()
	session, err := s.sessionApp.GetSession(ctx, key)
	if err != nil {
		// TODO(taekyeom) handle error
		return false, err
	}

	currentTime := utils.GetRequestTimeOrNow(ctx)

	if session.ExpireTime.Before(currentTime) {
		return false, nil
	}

	// Set userId key
	ctx = utils.SetUserId(ctx, session.UserId)
	r := c.Request().WithContext(ctx)
	c.SetRequest(r)

	return true, nil
}

func NewSessionMiddleware(sessionApp userSessionApp) sessionMiddleware {
	return sessionMiddleware{
		sessionApp: sessionApp,
	}
}
