package taxicall

import (
	"context"

	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/repository"
	"github.com/taco-labs/taco/go/service"
)

type userGetterInterface interface {
	GetUser(context.Context, string) (entity.User, error)
}

type driverGetterInterface interface {
	GetDriver(context.Context, string) (entity.Driver, error)
}

type taxicallApp struct {
	app.Transactor
	repository struct {
		driverLocation  repository.DriverLocationRepository
		taxiCallRequest repository.TaxiCallRepository
		event           repository.EventRepository
	}
	service struct {
		route        service.MapRouteService
		location     service.LocationService
		userGetter   userGetterInterface
		driverGetter driverGetterInterface
	}
}
