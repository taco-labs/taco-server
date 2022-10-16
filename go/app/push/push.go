package push

import (
	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/service"
)

type taxiCallPushApp struct {
	app.Transactor
	service struct {
		route        service.MapRouteService
		notification service.NotificationService
	}
}
