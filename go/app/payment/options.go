package payment

import (
	"errors"

	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/repository"
	"github.com/taco-labs/taco/go/service"
)

type paymentAppOption func(*paymentApp)

func WithTransactor(transactor app.Transactor) paymentAppOption {
	return func(pa *paymentApp) {
		pa.Transactor = transactor
	}
}

func WithPaymentRepository(repo repository.PaymentRepository) paymentAppOption {
	return func(pa *paymentApp) {
		pa.repository.payment = repo
	}
}

func WithEventRepository(repo repository.EventRepository) paymentAppOption {
	return func(pa *paymentApp) {
		pa.repository.event = repo
	}
}

func WithPaymentService(svc service.PaymentService) paymentAppOption {
	return func(pa *paymentApp) {
		pa.service.payment = svc
	}
}

func WithEventSubService(svc service.EventSubscriptionService) paymentAppOption {
	return func(pa *paymentApp) {
		pa.service.eventSub = svc
	}
}

func WithEventPubService(svc service.EventPublishService) paymentAppOption {
	return func(pa *paymentApp) {
		pa.service.eventPub = svc
	}
}

func WithWorkerPoolService(svc service.WorkerPoolService) paymentAppOption {
	return func(pa *paymentApp) {
		pa.service.workerPool = svc
	}
}

func (p paymentApp) validateApp() error {
	if p.Transactor == nil {
		return errors.New("user payment app need transactor")
	}

	if p.repository.payment == nil {
		return errors.New("user payment app need payment repository")
	}

	if p.repository.event == nil {
		return errors.New("user event app need event repository")
	}

	if p.service.payment == nil {
		return errors.New("user payment app need payment service")
	}

	if p.service.eventSub == nil {
		return errors.New("user payment app need event sub service")
	}

	if p.service.eventPub == nil {
		return errors.New("user payment app need event pub service")
	}

	if p.service.workerPool == nil {
		return errors.New("user payment app need worker pool service")
	}

	return nil
}

func NewPaymentApp(opts ...paymentAppOption) (*paymentApp, error) {
	app := &paymentApp{
		waitCh: make(chan struct{}),
	}

	for _, opt := range opts {
		opt(app)
	}

	return app, app.validateApp()
}
