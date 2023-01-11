package driversettlement

import (
	"errors"

	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/repository"
	"github.com/taco-labs/taco/go/service"
)

type driversettlementAppOption func(*driversettlementApp)

func WithTransactor(transactor app.Transactor) driversettlementAppOption {
	return func(da *driversettlementApp) {
		da.Transactor = transactor
	}
}

func WithSettlementRepository(repo repository.DriverSettlementRepository) driversettlementAppOption {
	return func(da *driversettlementApp) {
		da.repository.settlement = repo
	}
}

func WithEventRepository(repo repository.EventRepository) driversettlementAppOption {
	return func(da *driversettlementApp) {
		da.repository.event = repo
	}
}

func WithAnalyticsRepository(repo repository.AnalyticsRepository) driversettlementAppOption {
	return func(da *driversettlementApp) {
		da.repository.analytics = repo
	}
}

func WithSettlementAccountService(svc service.SettlementAccountService) driversettlementAppOption {
	return func(da *driversettlementApp) {
		da.service.settlementAccount = svc
	}
}

func WithMetricService(svc service.MetricService) driversettlementAppOption {
	return func(da *driversettlementApp) {
		da.service.metric = svc
	}
}

func (d driversettlementApp) validateApp() error {
	if d.Transactor == nil {
		return errors.New("driver settlement app need transactor")
	}

	if d.repository.settlement == nil {
		return errors.New("driver settlement app need settlement repository")
	}

	if d.repository.event == nil {
		return errors.New("driver settlement app need event repository")
	}

	if d.repository.analytics == nil {
		return errors.New("driver settlement app need analytics repository")
	}

	if d.service.settlementAccount == nil {
		return errors.New("driver settlement app need settlement account service")
	}

	if d.service.metric == nil {
		return errors.New("driver settlement app need metric service")
	}

	return nil
}

func NewDriverSettlementApp(opts ...driversettlementAppOption) (*driversettlementApp, error) {
	app := &driversettlementApp{}

	for _, opt := range opts {
		opt(app)
	}

	return app, app.validateApp()
}
