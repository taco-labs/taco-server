package push

import (
	"errors"

	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/repository"
	"github.com/taco-labs/taco/go/service"
)

type pushAppOption func(*taxiCallPushApp)

func WithTransactor(transactor app.Transactor) pushAppOption {
	return func(tcpa *taxiCallPushApp) {
		tcpa.Transactor = transactor
	}
}

func WithPushTokenRepository(repo repository.PushTokenRepository) pushAppOption {
	return func(tcpa *taxiCallPushApp) {
		tcpa.repository.pushToken = repo
	}
}

func WithRouteService(svc service.MapRouteService) pushAppOption {
	return func(tcpa *taxiCallPushApp) {
		tcpa.service.route = svc
	}
}

func WithNotificationService(svc service.NotificationService) pushAppOption {
	return func(tcpa *taxiCallPushApp) {
		tcpa.service.notification = svc
	}
}

func WithUserGetterService(svc userGetterInterface) pushAppOption {
	return func(tcpa *taxiCallPushApp) {
		tcpa.service.userGetter = svc
	}
}

func WithDriverGetterService(svc driverGetterInterface) pushAppOption {
	return func(tcpa *taxiCallPushApp) {
		tcpa.service.driverGetter = svc
	}
}

func (t taxiCallPushApp) validate() error {
	if t.Transactor == nil {
		return errors.New("taxi call push app need transactor")
	}

	if t.repository.pushToken == nil {
		return errors.New("taxi call push app need push token repository")
	}

	if t.service.route == nil {
		return errors.New("taxi call push app need route service")
	}

	if t.service.notification == nil {
		return errors.New("taxi call push app need notification service")
	}

	if t.service.userGetter == nil {
		return errors.New("taxi call push app need user getter")
	}

	if t.service.driverGetter == nil {
		return errors.New("taxi call push app need driver getter")
	}

	return nil
}

func NewPushApp(opts ...pushAppOption) (taxiCallPushApp, error) {
	app := taxiCallPushApp{}

	for _, opt := range opts {
		opt(&app)
	}

	return app, app.validate()
}
