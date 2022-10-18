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

func WithEventPublisherService(svc service.EventPublishService) taxicallAppOption {
	return func(ta *taxicallApp) {
		ta.service.eventPub = svc
	}
}

func WithEventSubscriberService(svc service.EventSubscriptionService) taxicallAppOption {
	return func(ta *taxicallApp) {
		ta.service.eventSub = svc
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

	if t.service.eventPub == nil {
		return errors.New("taxi call app needs event pub service")
	}

	if t.service.eventSub == nil {
		return errors.New("taxi call app needs event sub service")
	}

	return nil
}

func NewTaxicallApp(opts ...taxicallAppOption) (taxicallApp, error) {
	app := taxicallApp{
		waitCh: make(chan struct{}),
	}

	for _, opt := range opts {
		opt(&app)
	}

	return app, app.validateApp()
}
