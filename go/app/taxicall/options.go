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

func WithAnalyticsRepository(repo repository.AnalyticsRepository) taxicallAppOption {
	return func(ta *taxicallApp) {
		ta.repository.analytics = repo
	}
}

func WithMapService(svc service.MapService) taxicallAppOption {
	return func(ta *taxicallApp) {
		ta.service.mapService = svc
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

func WithPaymentAppService(svc paymentAppInterface) taxicallAppOption {
	return func(ta *taxicallApp) {
		ta.service.payment = svc
	}
}

func WithUserServiceRegionChecker(svc service.ServiceRegionChecker) taxicallAppOption {
	return func(ta *taxicallApp) {
		ta.service.userServiceRegionChecker = svc
	}
}

func WithMetricService(svc service.MetricService) taxicallAppOption {
	return func(ta *taxicallApp) {
		ta.service.metric = svc
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

	if t.repository.analytics == nil {
		return errors.New("taxi call app needs analytics repository")
	}

	if t.service.mapService == nil {
		return errors.New("taxi call app needs maps service")
	}

	if t.service.userGetter == nil {
		return errors.New("taxi call app needs user getter")
	}

	if t.service.driverGetter == nil {
		return errors.New("taxi call app needs driver getter")
	}

	if t.service.payment == nil {
		return errors.New("taxi call app needs payment service")
	}

	if t.service.userServiceRegionChecker == nil {
		return errors.New("taxi call app needs user service region checker")
	}

	if t.service.metric == nil {
		return errors.New("taxi call app needs metric service")
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
