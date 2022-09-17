package driver

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/taco-labs/taco/go/server"
	"github.com/labstack/echo/v4"
)

type driverServer struct {
	echo     *echo.Echo
	endpoint string
	port     int
	app      struct {
		driver driverApp
	}
	middlewares []echo.MiddlewareFunc
}

func (d *driverServer) initMiddleware() error {
	d.echo.Use(server.DefaultRequestTimeMiddelware.Process)

	for _, middleware := range d.middlewares {
		d.echo.Use(middleware)
	}

	return nil
}

func (d *driverServer) initController() error {
	d.echo.GET("/healthz", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	return nil
}

func (d *driverServer) Run(ctx context.Context) error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for {
			select {
			case <-c:
				fmt.Println("shutting down [User API] server... because of interrupt")
				d.echo.Shutdown(ctx)
				return
			case <-ctx.Done():
				fmt.Println("shutting down [User API] server... because of context cancel")
				d.echo.Shutdown(ctx)
				return
			}
		}
	}()
	if err := d.echo.Start(fmt.Sprintf("%s:%d", d.endpoint, d.port)); err != nil {
		return err
	}
	return nil
}

func NewDriverServer(opts ...driverServerOption) (driverServer, error) {
	e := echo.New()

	server := driverServer{
		echo:        e,
		middlewares: make([]echo.MiddlewareFunc, 0),
	}

	for _, opt := range opts {
		opt(&server)
	}

	if err := server.initMiddleware(); err != nil {
		return server, err
	}

	if err := server.initController(); err != nil {
		return server, err
	}

	return server, server.validate()
}

func (d driverServer) validate() error {
	if d.app.driver == nil {
		return fmt.Errorf("driver server need driver app")
	}
	return nil
}
