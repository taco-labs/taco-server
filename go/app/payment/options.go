package payment

import (
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

func WithPaymentRepository(repo repository.UserPaymentRepository) paymentAppOption {
	return func(pa *paymentApp) {
		pa.repository.payment = repo
	}
}

func WithPaymentService(svc service.CardPaymentService) paymentAppOption {
	return func(pa *paymentApp) {
		pa.service.payment = svc
	}
}

func WithEventSubService(svc service.EventSubscriptionService) paymentAppOption {
	return func(pa *paymentApp) {
		pa.service.eventSub = svc
	}
}

func WithWorkerPoolService(svc service.WorkerPoolService) paymentAppOption {
	return func(pa *paymentApp) {
		pa.service.workerPool = svc
	}
}
