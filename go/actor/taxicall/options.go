package taxicall

import (
	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/repository"
)

type actorOption func(*TaxiCallActorService)

func WithTransactor(transactor app.Transactor) actorOption {
	return func(tcas *TaxiCallActorService) {
		tcas.Transactor = transactor
	}
}

func WithUserRepository(repo repository.UserRepository) actorOption {
	return func(tcas *TaxiCallActorService) {
		tcas.repository.user = repo
	}
}

func WithDriverRepository(repo repository.DriverRepository) actorOption {
	return func(tcas *TaxiCallActorService) {
		tcas.repository.driver = repo
	}
}

func WithTaxiCallRequestRepository(repo repository.TaxiCallRepository) actorOption {
	return func(tcas *TaxiCallActorService) {
		tcas.repository.taxiCallRequest = repo
	}
}
