package user

import (
	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/repository"
	"github.com/taco-labs/taco/go/service"
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

func WithSmsVerificationRepository(repo repository.SmsVerificationRepository) userAppOption {
	return func(ua *userApp) {
		ua.repository.smsVerification = repo
	}
}

func WithTaxiCallRequestRepository(repo repository.TaxiCallRepository) userAppOption {
	return func(ua *userApp) {
		ua.repository.taxiCallRequest = repo
	}
}

func WithSessionService(app SessionInterface) userAppOption {
	return func(ua *userApp) {
		ua.service.session = app
	}
}

func WithCardPaymentService(svc service.CardPaymentService) userAppOption {
	return func(ua *userApp) {
		ua.service.payment = svc
	}
}

func WithSmsSenderService(svc service.SmsSenderService) userAppOption {
	return func(ua *userApp) {
		ua.service.smsSender = svc
	}
}

func WithMapRouteService(svc service.MapRouteService) userAppOption {
	return func(ua *userApp) {
		ua.service.route = svc
	}
}

func WithLocationService(svc service.LocationService) userAppOption {
	return func(ua *userApp) {
		ua.service.location = svc
	}
}
