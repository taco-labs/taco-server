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

func WithSmsVerificationRepository(repo repository.SmsVerificationRepository) userAppOption {
	return func(ua *userApp) {
		ua.repository.smsVerification = repo
	}
}

func WithAnalyticsRepository(repo repository.AnalyticsRepository) userAppOption {
	return func(ua *userApp) {
		ua.repository.analytics = repo
	}
}

func WithSessionService(app sessionServiceInterface) userAppOption {
	return func(ua *userApp) {
		ua.service.session = app
	}
}

func WithSmsSenderService(svc service.SmsSenderService) userAppOption {
	return func(ua *userApp) {
		ua.service.smsSender = svc
	}
}

func WithMapService(svc service.MapService) userAppOption {
	return func(ua *userApp) {
		ua.service.mapService = svc
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

func WithUserPaymentService(svc userPaymentInterface) userAppOption {
	return func(ua *userApp) {
		ua.service.userPayment = svc
	}
}

func WithDriverAppService(svc driverAppInterface) userAppOption {
	return func(ua *userApp) {
		ua.service.driver = svc
	}
}

func (u userApp) validateApp() error {
	if u.Transactor == nil {
		return errors.New("user app need transator")
	}

	if u.repository.user == nil {
		return errors.New("user app need user repository")
	}

	if u.repository.smsVerification == nil {
		return errors.New("user app need sms verification repository")
	}

	if u.repository.analytics == nil {
		return errors.New("user app need analytics repository")
	}

	if u.service.session == nil {
		return errors.New("user app need user session repository")
	}

	if u.service.smsSender == nil {
		return errors.New("user app need sms sender service")
	}

	if u.service.mapService == nil {
		return errors.New("user app need map service")
	}

	if u.service.push == nil {
		return errors.New("user app need push service")
	}

	if u.service.taxiCall == nil {
		return errors.New("user app need taxi call service")
	}

	if u.service.userPayment == nil {
		return errors.New("user app need user payment service")
	}

	if u.service.driver == nil {
		return errors.New("user app need driver app service")
	}

	return nil
}
