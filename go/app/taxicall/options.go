package taxicall

import (
	"errors"

	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/repository"
	"github.com/taco-labs/taco/go/service"
)

type taxicallAppOption func(*taxicallApp)

func WithTransactor(transactor app.Transactor) taxicallAppOption {
	return func(ta *taxicallApp) {
		ta.Transactor = transactor
	}
}

func WithDriverLocationRepository(repo repository.DriverLocationRepository) taxicallAppOption {
	return func(ta *taxicallApp) {
		ta.repository.driverLocation = repo
	}
}

func WithTaxiCallRequestRepository(repo repository.TaxiCallRepository) taxicallAppOption {
	return func(ta *taxicallApp) {
		ta.repository.taxiCallRequest = repo
	}
}

func WithEventRepository(repo repository.EventRepository) taxicallAppOption {
	return func(ta *taxicallApp) {
		ta.repository.event = repo
	}
}

func WithRouteServie(svc service.MapRouteService) taxicallAppOption {
	return func(ta *taxicallApp) {
		ta.service.route = svc
	}
}

func WithLocationService(svc service.LocationService) taxicallAppOption {
	return func(ta *taxicallApp) {
		ta.service.location = svc
	}
}

func WithUserGetterService(svc userGetterInterface) taxicallAppOption {
	return func(ta *taxicallApp) {
		ta.service.userGetter = svc
	}
}

func WithDriverGetterService(svc driverGetterInterface) taxicallAppOption {
	return func(ta *taxicallApp) {
		ta.service.driverGetter = svc
	}
}

func (t taxicallApp) validateApp() error {
	if t.Transactor == nil {
		return errors.New("taxi call app needs transactor ")
	}

	if t.repository.driverLocation == nil {
		return errors.New("taxi call app needs driver location repository")
	}

	if t.repository.taxiCallRequest == nil {
		return errors.New("taxi call app needs taxi call repository")
	}

	if t.repository.event == nil {
		return errors.New("taxi call app needs event repository")
	}

	if t.service.route == nil {
		return errors.New("taxi call app needs route service")
	}

	if t.service.location == nil {
		return errors.New("taxi call app needs location service")
	}

	if t.service.userGetter == nil {
		return errors.New("taxi call push app need user getter")
	}

	if t.service.driverGetter == nil {
		return errors.New("taxi call push app need driver getter")
	}

	return nil
}

func NewTaxicallApp(opts ...taxicallAppOption) (taxicallApp, error) {
	app := taxicallApp{}

	for _, opt := range opts {
		opt(&app)
	}

	return app, app.validateApp()
}
