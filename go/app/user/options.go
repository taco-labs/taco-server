package user

import (
	"errors"

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

func WithSessionService(app sessionServiceInterface) userAppOption {
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

func WithPushService(svc pushServiceInterface) userAppOption {
	return func(ua *userApp) {
		ua.service.push = svc
	}
}

func WithTaxiCallService(svc taxiCallInterface) userAppOption {
	return func(ua *userApp) {
		ua.service.taxiCall = svc
	}
}

func (u userApp) validateApp() error {
	if u.Transactor == nil {
		return errors.New("user app need transator")
	}

	if u.repository.user == nil {
		return errors.New("user app need user repository")
	}

	if u.repository.payment == nil {
		return errors.New("user app need user payment repository")
	}

	if u.repository.smsVerification == nil {
		return errors.New("user app need sms verification repository")
	}

	if u.service.session == nil {
		return errors.New("user app need user session repository")
	}

	if u.service.payment == nil {
		return errors.New("user app need card payment service")
	}

	if u.service.smsSender == nil {
		return errors.New("user app need sms sender service")
	}

	if u.service.route == nil {
		return errors.New("user app need map route service")
	}

	if u.service.location == nil {
		return errors.New("user app need location service")
	}

	if u.service.push == nil {
		return errors.New("user app need push service")
	}

	if u.service.taxiCall == nil {
		return errors.New("user app need taxi call service")
	}

	return nil
}
