package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/labstack/echo/v4"
)

type userServer struct {
	endpoint string
	port     int
	echo     *echo.Echo
	app      struct {
		user UserApp
	}
}

func (u *userServer) initMiddleware() error {
	u.echo.Use(defaultRequestTimeMiddelware.Process)

	return nil
}

func (u *userServer) initController() error {
	u.echo.GET("/healthz", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	u.echo.POST("/signup", u.Signup)
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

func NewUserServer(endpoint string, port int, userApp UserApp) (userServer, error) {
	e := echo.New()

	server := userServer{
		endpoint: endpoint,
		port:     port,
		echo:     e,
	}

	server.app.user = userApp

	if err := server.initMiddleware(); err != nil {
		return userServer{}, err
	}

	if err := server.initController(); err != nil {
		return userServer{}, err
	}

	return server, nil
}
