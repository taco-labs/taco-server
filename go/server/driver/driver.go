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
	driverGroup.GET("/:driverId/image_urls", d.GetDriverImageUrls)
	driverGroup.PUT("/:driverId/on_duty", d.UpdateOnDuty)
	driverGroup.PUT("/:driverId/location", d.UpdateDriverLocation)
	driverGroup.POST("/:driverId/settlement_account", d.RegisterDriverSettlementAccount)
	driverGroup.GET("/:driverId/settlement_account", d.GetDriverSettlemtnAccount)
	driverGroup.PUT("/:driverId/settlement_account", d.UpdateDriverSettlemtnAccount)
	driverGroup.GET("/:driverId/expected_settlement", d.GetExpectedDriverSetttlement)
	driverGroup.GET("/:driverId/settlement", d.ListDriverSettlementHistory)
	driverGroup.POST("/:driverId/settlement_request", d.RequestDriverSettlementTransfer)
	driverGroup.GET("/:driverId/taxicall_latest", d.GetLatestTaxiCallRequest)
	driverGroup.GET("/:driverId/taxicall", d.ListTaxiCallRequest)
	driverGroup.GET("/:driverId/ticket_latest", d.LatestTaxiCallTicket)
	driverGroup.PUT("/:driverId/deny_taxi_call_tag/:tagId", d.AddDriverDenyTaxiCallTag)
	driverGroup.DELETE("/:driverId/deny_taxi_call_tag/:tagId", d.DeleteDriverDenyTaxiCallTag)
	driverGroup.GET("/:driverId/deny_taxi_call_tag", d.ListDriverDenyTaxiCallTag)

	driverGroup.GET("/:driverId/car_profile", d.ListDriverCarProfile)
	driverGroup.PUT("/:driverId/car_profile/:carProfileId", d.SelectDriverCarProfile)

	taxiCallGroup := d.echo.Group("/taxicall")
	taxiCallGroup.PUT("/ticket/:ticketId", d.AcceptTaxiCallRequest)
	taxiCallGroup.DELETE("/ticket/:ticketId", d.RejectTaxiCallRequest)

	// TODO (taekyeom) Proper url pattern..
	taxiCallGroup.DELETE("/:taxiCallRequestId", d.CancelTaxiCallRequest)
	taxiCallGroup.PUT("/:taxiCallRequestId/to_arrival", d.DriverToArrival)
	taxiCallGroup.PUT("/:taxiCallRequestId/done", d.DoneTaxiCallRequest)

	carProfileGroup := d.echo.Group("/car_profile")
	carProfileGroup.POST("", d.AddDriverCarProfile)
	carProfileGroup.GET("/:carProfileId", d.GetDriverCarProfile)
	carProfileGroup.PUT("/:carProfileId", d.UpdateDriverCarProfile)
	carProfileGroup.DELETE("/:carProfileId", d.DeleteDriverCarProfile)

	d.echo.GET("/service_region", d.ListAvailableServiceRegion)

	// TODO (taekyeom) 별도로 추상화 된 구현으로 빼야 함
	d.echo.GET("/download", func(c echo.Context) error {
		return c.Redirect(http.StatusSeeOther, "https://play.google.com/store/apps/details?id=com.tacolabs.taco.driver")
	})

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
