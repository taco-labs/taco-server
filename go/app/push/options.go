package push

import (
	"errors"

	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/service"
)

type pushAppOption func(*taxiCallPushApp)

func WithTransactor(transactor app.Transactor) pushAppOption {
	return func(tcpa *taxiCallPushApp) {
		tcpa.Transactor = transactor
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

func (t taxiCallPushApp) validate() error {
	if t.service.route == nil {
		return errors.New("taxi call push app need route service")
	}

	if t.service.notification == nil {
		return errors.New("taxi call push app need notification service")
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
