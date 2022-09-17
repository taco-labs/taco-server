package usersession

import (
	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/repository"
)

type userSessionOption func(*userSessionApp)

func WithTransactor(transcator app.Transactor) userSessionOption {
	return func(usa *userSessionApp) {
		usa.Transactor = transcator
	}
}

func WithUserSessionRepository(repo repository.UserSessionRepository) userSessionOption {
	return func(usa *userSessionApp) {
		usa.repository.session = repo
	}
}
