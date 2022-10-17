package driver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/taco-labs/taco/go/server"
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

	driverGroup := d.echo.Group("/driver")
	driverGroup.POST("/signin/sms/request", d.SmsVerificationRequest)
	driverGroup.POST("/signin/sms/verify", d.SmsSignin)
	driverGroup.POST("/signup", d.Signup)
	driverGroup.GET("/:driverId", d.GetDriver)
	driverGroup.PUT("/:driverId", d.UpdateDriver)
	driverGroup.PUT("/:driverId/on_duty", d.UpdateOnDuty)
	driverGroup.PUT("/:driverId/location", d.UpdateDriverLocation)
	driverGroup.POST("/:driverId/settlement_account", d.RegisterDriverSettlementAccount)
	driverGroup.GET("/:driverId/settlement_account", d.GetDriverSettlemtnAccount)
	driverGroup.PUT("/:driverId/settlement_account", d.UpdateDriverSettlemtnAccount)
	driverGroup.GET("/:driverId/taxicall_latest", d.GetLatestTaxiCallRequest)
	driverGroup.GET("/:driverId/taxicall", d.ListTaxiCallRequest)

	taxiCallGroup := d.echo.Group("/taxicall")
	taxiCallGroup.PUT("/ticket/:ticketId", d.AcceptTaxiCallRequest)
	taxiCallGroup.DELETE("/ticket/:ticketId", d.RejectTaxiCallRequest)

	// TODO (taekyeom) Proper url pattern..
	taxiCallGroup.PUT("/:taxiCallRequestId/to_arrival", d.DriverToArrival)
	taxiCallGroup.PUT("/:taxiCallRequestId/done", d.DoneTaxiCallRequest)

	return nil
}

func (d *driverServer) Run(ctx context.Context) error {
	if err := d.echo.Start(fmt.Sprintf("%s:%d", d.endpoint, d.port)); err != nil {
		return err
	}
	return nil
}

func (d driverServer) Stop(ctx context.Context) error {
	fmt.Println("shutting down [Driver API] server...")
	return d.echo.Shutdown(ctx)
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
