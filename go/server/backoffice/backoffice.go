package backoffice

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/request"
)

type driverApp interface {
	GetDriver(context.Context, string) (entity.Driver, error)
	DeleteDriver(context.Context, string) error
	ActivateDriver(context.Context, string) error

	// TODO(taekyeom) Must remove before production
	DriverToArrival(context.Context, string) error
	ForceAcceptTaxiCallRequest(context.Context, string, string) (entity.DriverLatestTaxiCallRequest, error)
	DoneTaxiCallRequest(context.Context, request.DoneTaxiCallRequest) error

	ListNonActivatedDriver(context.Context, request.ListNonActivatedDriverRequest) ([]entity.DriverDto, string, error)

	GetDriverSettlementAccount(ctx context.Context, driverId string) (entity.DriverSettlementAccount, error)
}

type taxicallApp interface {
	ListDriverTaxiCallContextInRadius(ctx context.Context, req request.ListDriverTaxiCallContextInRadiusRequest) ([]entity.DriverTaxiCallContextWithInfo, error)
}

type userApp interface {
	GetUser(context.Context, string) (entity.User, error)
	DeleteUser(context.Context, string) error
}

type backofficeServer struct {
	echo     *echo.Echo
	endpoint string
	port     int
	app      struct {
		driver   driverApp
		user     userApp
		taxicall taxicallApp
	}
	middlewares []echo.MiddlewareFunc
}

func (b *backofficeServer) initMiddleware() error {
	for _, middleware := range b.middlewares {
		b.echo.Use(middleware)
	}

	return nil
}

func (b *backofficeServer) initController() error {
	b.echo.GET("/healthz", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	driverGroup := b.echo.Group("/driver")
	driverGroup.GET("/non_active", b.ListNonActivatedDriver)
	driverGroup.GET("/:driverId", b.GetDriver)
	driverGroup.DELETE("/:driverId", b.DeleteDriver)
	driverGroup.PUT("/:driverId/activate", b.ActivateDriver)
	driverGroup.PUT("/:driverId/force_accept/:taxiCallRequestId", b.ForceAcceptTaxiCallRequest)
	driverGroup.PUT("/:driverId/to_arrival/:taxiCallRequestId", b.DriverToArrival)
	driverGroup.PUT("/:driverId/done/:taxiCallRequestId", b.DoneTaxiCallRequest)
	driverGroup.GET("/:driverId/settlement_account", b.GetDriverSettlementAccount)

	userGroup := b.echo.Group("/user")
	userGroup.GET("/:userId", b.GetUser)
	userGroup.DELETE("/:userId", b.DeleteUser)

	taxicallGroup := b.echo.Group("/taxicall")
	taxicallGroup.GET("/available_drivers", b.ListDriverTaxiCallContextInRadius)

	return nil
}

func (b backofficeServer) validate() error {
	if b.app.driver == nil {
		return errors.New("backoffice server need driver app")
	}

	if b.app.user == nil {
		return errors.New("backoffice server need user app")
	}

	if b.app.taxicall == nil {
		return errors.New("backoffice server need taxi call app")
	}

	return nil
}

func (b *backofficeServer) Run(ctx context.Context) error {
	if err := b.echo.Start(fmt.Sprintf("%s:%d", b.endpoint, b.port)); err != nil {
		return err
	}
	return nil
}

func (b *backofficeServer) Stop(ctx context.Context) error {
	fmt.Println("shutting down [Backoffice API] server...")
	return b.echo.Shutdown(ctx)
}

func NewBackofficeServer(opts ...backofficeOption) (backofficeServer, error) {
	e := echo.New()

	server := backofficeServer{
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
