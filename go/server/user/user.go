package user

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/taco-labs/taco/go/server"
)

type userServer struct {
	echo     *echo.Echo
	endpoint string
	port     int
	app      struct {
		user UserApp
	}
	middlewares []echo.MiddlewareFunc
}

func (u *userServer) initMiddleware() error {
	u.echo.Use(server.DefaultRequestTimeMiddelware.Process)

	for _, middleware := range u.middlewares {
		u.echo.Use(middleware)
	}

	return nil
}

func (u *userServer) initController() error {
	u.echo.GET("/healthz", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	userGroup := u.echo.Group("/user")
	userGroup.POST("/signin/sms/request", u.SmsVerificationRequest)
	userGroup.POST("/signin/sms/verify", u.SmsSignin)
	userGroup.POST("/signup", u.Signup)
	userGroup.GET("/:userId", u.GetUser)
	userGroup.PUT("/:userId", u.UpdateUser)
	userGroup.GET("/:userId/payment", u.ListUserPayment)
	userGroup.POST("/:userId/payment/:paymentId/default", u.UpdateDefaultPayment)
	userGroup.GET("/:userId/taxicall_latest", u.GetLatestTaxiCallRequest)
	userGroup.GET("/:userId/taxicall", u.ListTaxiCallRequest)

	paymentGroup := u.echo.Group("/payment")
	paymentGroup.POST("", u.RegisterUserPayment)
	paymentGroup.DELETE("/:paymentId", u.DeleteUserPayment)
	paymentGroup.PUT("/:paymentId/recovery", u.TryRecoverUserPayment)

	taxiCallGroup := u.echo.Group("/taxicall")
	taxiCallGroup.GET("/tags", u.ListTags)
	taxiCallGroup.POST("", u.CreateTaxiCallRequest)
	taxiCallGroup.DELETE("/:taxiCallRequestId", u.CancelTaxiCallRequest)

	locationGroup := u.echo.Group("/location")
	locationGroup.GET("/address", u.GetAddress)
	locationGroup.GET("/search", u.SearchLocation)
	return nil
}

func (u *userServer) Run(ctx context.Context) error {
	if err := u.echo.Start(fmt.Sprintf("%s:%d", u.endpoint, u.port)); err != nil {
		return err
	}
	return nil
}

func (u *userServer) Stop(ctx context.Context) error {
	fmt.Println("shutting down [User API] server...")
	return u.echo.Shutdown(ctx)
}

func NewUserServer(opts ...userServerOption) (userServer, error) {
	e := echo.New()

	server := userServer{
		echo:        e,
		middlewares: make([]echo.MiddlewareFunc, 0),
	}

	for _, opt := range opts {
		opt(&server)
	}

	if err := server.initMiddleware(); err != nil {
		return userServer{}, err
	}

	if err := server.initController(); err != nil {
		return userServer{}, err
	}

	return server, nil
}
