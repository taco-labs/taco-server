package backoffice

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var skipSet = map[string]struct{}{
	"/healthz": {},
}

type sessionMiddleware struct {
	secretKey string
}

func (s sessionMiddleware) Get() echo.MiddlewareFunc {
	return middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		Skipper: s.skipper,
		Validator: func(auth string, c echo.Context) (bool, error) {
			return auth == s.secretKey, nil
		},
	})
}

func (s sessionMiddleware) skipper(c echo.Context) bool {
	_, ok := skipSet[c.Path()]
	return ok
}

func NewSessionMiddleware(secretKey string) sessionMiddleware {
	return sessionMiddleware{secretKey}
}
