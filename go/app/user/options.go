package user

import (
	"github.com/ktk1012/taco/go/app"
	"github.com/ktk1012/taco/go/repository"
	"github.com/ktk1012/taco/go/service"
)

// User App
type userAppOption func(*userApp)

func WithTransactor(transactor app.Transactor) userAppOption {
	return func(ua *userApp) {
		ua.Transactor = transactor
	}
}

func WithUserRepository(repo repository.UserRepository) userAppOption {
	return func(ua *userApp) {
		ua.repository.user = repo
	}
}

func WithUserPaymentRepository(repo repository.UserPaymentRepository) userAppOption {
	return func(ua *userApp) {
		ua.repository.payment = repo
	}
}

func WithSessionService(app SessionInterface) userAppOption {
	return func(ua *userApp) {
		ua.service.session = app
	}
}

func WithUserIdentityService(svc service.UserIdentityService) userAppOption {
	return func(ua *userApp) {
		ua.service.userIdentity = svc
	}
}

func WithCardPaymentService(svc service.CardPaymentService) userAppOption {
	return func(ua *userApp) {
		ua.service.payment = svc
	}
}
