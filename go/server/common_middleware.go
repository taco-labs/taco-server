package server

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/taco-labs/taco/go/utils"
	"go.uber.org/zap"
)

var (
	DefaultRequestTimeMiddelware = newRequestTimeMiddleware(defaultTimer{})
)

type timer interface {
	Now() time.Time
}

type defaultTimer struct{}

func (d defaultTimer) Now() time.Time {
	return time.Now()
}

type requestTimeMiddleware struct {
	timer timer
}

func (r requestTimeMiddleware) Process(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestTime := r.timer.Now()
		ctx := c.Request().Context()

		ctx = utils.SetRequestTime(ctx, requestTime)

		r := c.Request().WithContext(ctx)
		c.SetRequest(r)
		return next(c)
	}
}

func newRequestTimeMiddleware(timer timer) requestTimeMiddleware {
	return requestTimeMiddleware{timer}
}

type loggerMiddleware struct {
	logger *zap.Logger
}

func (r loggerMiddleware) Process(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		ctx = utils.SetLogger(ctx, r.logger)

		r := c.Request().WithContext(ctx)
		c.SetRequest(r)
		return next(c)
	}
}

func NewLoggerMiddleware(logger *zap.Logger) loggerMiddleware {
	return loggerMiddleware{
		logger: logger,
	}
}
