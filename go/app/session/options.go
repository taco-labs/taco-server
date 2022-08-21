package session

import (
	"github.com/ktk1012/taco/go/app"
	"github.com/ktk1012/taco/go/repository"
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
