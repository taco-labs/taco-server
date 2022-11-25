package driversettlement

import (
	"errors"

	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/repository"
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

func (d driversettlementApp) validateApp() error {
	if d.Transactor == nil {
		return errors.New("driver settlement app need transactor")
	}

	if d.repository.settlement == nil {
		return errors.New("driver settlement app need settlement repository")
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
