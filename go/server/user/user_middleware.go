package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/domain/value/analytics"
	"github.com/taco-labs/taco/go/repository"
	"github.com/taco-labs/taco/go/server"
	"github.com/taco-labs/taco/go/utils"
	"github.com/uptrace/bun"
)

var skipSet = map[string]struct{}{
	"/user/signin/sms/request": {},
	"/user/signin/sms/verify":  {},
	"/user/signup":             {},
	"/healthz":                 {},
	"/download":                {},

	// TODO (taekyeom) skip set도 외부 extension이 제어 가능하도록 개선 필요
	"/payment/payple/result_callback":      {},
	"/payment/payple/register_success":     {},
	"/payment/payple/register_failure":     {},
	"/payment/payple/register_cancel":      {},
	"/payment/payple/transaction_callback": {},
}

type userSessionApp interface {
	GetSession(context.Context, string) (entity.UserSession, error)
	UpdateSession(context.Context, entity.UserSession) error
}

type sessionMiddleware struct {
	sessionApp userSessionApp
}

func (s sessionMiddleware) Get() echo.MiddlewareFunc {
	return middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		Skipper:   s.skipper,
		Validator: s.validateSession,
		KeyLookup: fmt.Sprintf("header:%s,query:apiKey", echo.HeaderAuthorization),
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
	session, err := s.sessionApp.GetSession(ctx, key)
	if errors.Is(err, value.ErrNotFound) {
		return false, value.ErrUnAuthenticated
	}
	if err != nil {
		// TODO(taekyeom) handle error
		return false, err
	}

	currentTime := utils.GetRequestTimeOrNow(ctx)

	if session.ExpireTime.Before(currentTime) {
		return false, value.ErrSessionExpired
	}

	// If we need to extend expiration time, update it
	if session.ExpireTime.Sub(currentTime) < (time.Hour*24)*7 {
		session.ExpireTime = currentTime.AddDate(0, 1, 0)
		if err := s.sessionApp.UpdateSession(ctx, session); err != nil {
			return false, err
		}
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

func UserIdChecker(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		requestUserId := c.Param("userId")

		if requestUserId != "" {
			userId := utils.GetUserId(ctx)
			if requestUserId != userId {
				return server.ToResponse(c, fmt.Errorf("unauthorized access to user resource:%w", value.ErrUnAuthorized))
			}
		}

		return next(c)
	}
}

type userAccessAnalyticsMiddleware struct {
	app.Transactor
	analytics repository.AnalyticsRepository
}

func NewUserAnalyticsMiddleware(transactor app.Transactor, analyticsRepository repository.AnalyticsRepository) *userAccessAnalyticsMiddleware {
	return &userAccessAnalyticsMiddleware{
		Transactor: transactor,
		analytics:  analyticsRepository,
	}
}

func (u userAccessAnalyticsMiddleware) Process(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		// TODO (taekyeom) user id 를 가져오기 위해서 일단은 skipset을 참조하도록 함
		if _, ok := skipSet[c.Path()]; ok {
			return next(c)
		}

		if err = next(c); err != nil {
			c.Error(err)
		}

		ctx := c.Request().Context()
		requestTime := utils.GetRequestTimeOrNow(ctx)
		userId := utils.GetUserId(ctx)
		path := c.Path()
		method := c.Request().Method

		err = u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
			if err := u.analytics.Create(ctx, i, entity.NewAnalytics(requestTime, analytics.UserAccessPayload{
				UserId: userId,
				Path:   path,
				Method: method,
			})); err != nil {
				return fmt.Errorf("middleware.user.access: error from analytics: %w", err)
			}

			return nil
		})

		return
	}
}
