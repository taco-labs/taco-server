package driver

import (
	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/repository"
	"github.com/taco-labs/taco/go/service"
)

type driverAppOption func(*driverApp)

func WithTransactor(transactor app.Transactor) driverAppOption {
	return func(da *driverApp) {
		da.Transactor = transactor
	}
}

func WithDriverRepository(repo repository.DriverRepository) driverAppOption {
	return func(da *driverApp) {
		da.repository.driver = repo
	}
}

func WithDriverLocationRepository(repo repository.DriverLocationRepository) driverAppOption {
	return func(da *driverApp) {
		da.repository.driverLocation = repo
	}
}

func WithSmsVerificationRepository(repo repository.SmsVerificationRepository) driverAppOption {
	return func(da *driverApp) {
		da.repository.smsVerification = repo
	}
}

func WithSettlementAccountRepository(repo repository.DriverSettlementAccountRepository) driverAppOption {
	return func(da *driverApp) {
		da.repository.settlementAccount = repo
	}
}

func WithTaxiCallRequestRepository(repo repository.TaxiCallRepository) driverAppOption {
	return func(da *driverApp) {
		da.repository.taxiCallRequest = repo
	}
}

func WithSessionService(svc sessionInterface) driverAppOption {
	return func(da *driverApp) {
		da.service.session = svc
	}
}

func WithSmsSenderService(svc service.SmsSenderService) driverAppOption {
	return func(da *driverApp) {
		da.service.smsSender = svc
	}
}

func WithFileUploadService(svc service.FileUploadService) driverAppOption {
	return func(da *driverApp) {
		da.service.fileUpload = svc
	}
}

func WithPushService(svc pushInterface) driverAppOption {
	return func(da *driverApp) {
		da.service.push = svc
	}
}
