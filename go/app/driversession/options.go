package driversession

import (
	"github.com/ktk1012/taco/go/app"
	"github.com/ktk1012/taco/go/repository"
)

type driverSessionOption func(*driverSessionApp)

func WithTransactor(transactor app.Transactor) driverSessionOption {
	return func(dsa *driverSessionApp) {
		dsa.Transactor = transactor
	}
}

func WithDriverSessionRepository(repo repository.DriverSessionRepository) driverSessionOption {
	return func(dsa *driverSessionApp) {
		dsa.repository.session = repo
	}
}
