package taxicall

import (
	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/repository"
	"github.com/taco-labs/taco/go/service"
)

type taxicallApp struct {
	app.Transactor
	repository struct {
		driverLocation  repository.DriverLocationRepository
		taxiCallRequest repository.TaxiCallRepository
		event           repository.EventRepository
	}
	service struct {
		route    service.MapRouteService
		location service.LocationService
	}
}
