package backoffice

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type sessionMiddleware struct {
	secretKey string
}

func (s sessionMiddleware) Get() echo.MiddlewareFunc {
	return middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		Validator: func(auth string, c echo.Context) (bool, error) {
			return auth == s.secretKey, nil
		},
	})
}

func NewSessionMiddleware(secretKey string) sessionMiddleware {
	return sessionMiddleware{secretKey}
}
