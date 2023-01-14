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
	userGroup.GET("/:userId/payment_point", u.GetUserPaymentPoint)
	userGroup.GET("/:userId/taxicall_latest", u.GetLatestTaxiCallRequest)
	userGroup.GET("/:userId/taxicall", u.ListTaxiCallRequest)

	paymentGroup := u.echo.Group("/payment")
	paymentGroup.DELETE("/:paymentId", u.DeleteUserPayment)
	paymentGroup.PUT("/:paymentId/recovery", u.TryRecoverUserPayment)

	taxiCallGroup := u.echo.Group("/taxicall")
	taxiCallGroup.GET("/tags", u.ListTags)
	taxiCallGroup.POST("", u.CreateTaxiCallRequest)
	taxiCallGroup.DELETE("/:taxiCallRequestId", u.CancelTaxiCallRequest)

	locationGroup := u.echo.Group("/location")
	locationGroup.GET("/address", u.GetAddress)
	locationGroup.GET("/search", u.SearchLocation)

	u.echo.GET("/service_region", u.ListAvailableServiceRegion)

	// TODO (taekyeom) 별도로 추상화 된 구현으로 빼야 함
	u.echo.GET("/download", func(c echo.Context) error {
		c.Response().Header().Add("Access-Control-Allow-Origin", "*")
		return c.Redirect(http.StatusSeeOther, "https://www.taco-labs.com")
	})

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
