package driver

import (
	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/repository"
)

type driverOption func(*driverApp)

func WithTransactor(transactor app.Transactor) driverOption {
	return func(da *driverApp) {
		da.Transactor = transactor
	}
}

func WithDriverRepository(repo repository.DriverRepository) driverOption {
	return func(da *driverApp) {
		da.repository.driver = repo
	}
}

func WithDriverLocationRepository(repo repository.DriverLocationRepository) driverOption {
	return func(da *driverApp) {
		da.repository.driverLocation = repo
	}
}

func WithSettlementAccountRepository(repo repository.DriverSettlementAccountRepository) driverOption {
	return func(da *driverApp) {
		da.repository.settlementAccount = repo
	}
}

func WithSessionService(svc sessionInterface) driverOption {
	return func(da *driverApp) {
		da.service.session = svc
	}
}
