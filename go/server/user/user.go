package user

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"

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
	userGroup.POST("/signin/sms/verify", u.SmsSingin)
	userGroup.POST("/signup", u.Signup)
	userGroup.GET("/:userId", u.GetUser)
	userGroup.PUT("/:userId", u.UpdateUser)
	userGroup.DELETE("/:userId", u.DeleteUser)
	userGroup.GET("/:userId/payment", u.ListCardPayment)
	userGroup.GET("/:userId/taxicall", u.ListTaxiCallRequest)

	paymentGroup := u.echo.Group("/payment")
	paymentGroup.POST("", u.RegisterCardPayment)
	paymentGroup.DELETE("/:paymentId", u.DeleteCardPayment)
	return nil
}

func (u *userServer) Run(ctx context.Context) error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for {
			select {
			case <-c:
				fmt.Println("shutting down [User API] server... because of interrupt")
				u.echo.Shutdown(ctx)
				return
			case <-ctx.Done():
				fmt.Println("shutting down [User API] server... because of context cancel")
				u.echo.Shutdown(ctx)
				return
			}
		}
	}()
	if err := u.echo.Start(fmt.Sprintf("%s:%d", u.endpoint, u.port)); err != nil {
		return err
	}
	return nil
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
