package driver

import (
	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/repository"
	"github.com/taco-labs/taco/go/service"
)

type driverOption func(*driverApp)

func WithTransactor(transactor app.Transactor) driverOption {
	return func(da *driverApp) {
		da.Transactor = transactor
	}
}

func WithDriverRepository(repo repository.DriverRepository) driverOption {
	return func(da *driverApp) {
		da.repository.driver = repo
	}
}

func WithDriverLocationRepository(repo repository.DriverLocationRepository) driverOption {
	return func(da *driverApp) {
		da.repository.driverLocation = repo
	}
}

func WithSmsVerificationRepository(repo repository.SmsVerificationRepository) driverOption {
	return func(da *driverApp) {
		da.repository.smsVerification = repo
	}
}

func WithSettlementAccountRepository(repo repository.DriverSettlementAccountRepository) driverOption {
	return func(da *driverApp) {
		da.repository.settlementAccount = repo
	}
}

func WithSessionService(svc sessionInterface) driverOption {
	return func(da *driverApp) {
		da.service.session = svc
	}
}

func WithSmsSenderService(svc service.SmsSenderService) driverOption {
	return func(da *driverApp) {
		da.service.smsSender = svc
	}
}

func WithFileUploadService(svc service.FileUploadService) driverOption {
	return func(da *driverApp) {
		da.service.fileUpload = svc
	}
}
