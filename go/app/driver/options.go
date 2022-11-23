package driver

import (
	"errors"

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

func WithEventRepository(repo repository.EventRepository) driverAppOption {
	return func(da *driverApp) {
		da.repository.event = repo
	}
}

func WithSessionService(svc sessionServiceInterface) driverAppOption {
	return func(da *driverApp) {
		da.service.session = svc
	}
}

func WithSmsSenderService(svc service.SmsSenderService) driverAppOption {
	return func(da *driverApp) {
		da.service.smsSender = svc
	}
}

func WithPushService(svc pushServiceInterface) driverAppOption {
	return func(da *driverApp) {
		da.service.push = svc
	}
}

func WithTaxiCallService(svc driverTaxiCallInterface) driverAppOption {
	return func(da *driverApp) {
		da.service.taxiCall = svc
	}
}

func WithImageUrlService(svc service.ImageUrlService) driverAppOption {
	return func(da *driverApp) {
		da.service.imageUrl = svc
	}
}

func WithSettlementAccountService(svc service.SettlementAccountService) driverAppOption {
	return func(da *driverApp) {
		da.service.settlementAccount = svc
	}
}

func (d driverApp) validateApp() error {
	if d.Transactor == nil {
		return errors.New("driver app need transactor")
	}

	if d.repository.driver == nil {
		return errors.New("driver app need driver repository")
	}

	if d.repository.smsVerification == nil {
		return errors.New("driver app need sms verification repository")
	}

	if d.repository.settlementAccount == nil {
		return errors.New("driver app need settlement account repository")
	}

	if d.repository.event == nil {
		return errors.New("driver app need event repository")
	}

	if d.service.session == nil {
		return errors.New("driver app need driver session service")
	}

	if d.service.smsSender == nil {
		return errors.New("driver app need sms sender service")
	}

	if d.service.push == nil {
		return errors.New("driver app need push service")
	}

	if d.service.taxiCall == nil {
		return errors.New("driver app need taxi call service")
	}

	if d.service.imageUrl == nil {
		return errors.New("driver app need image url service")
	}

	if d.service.settlementAccount == nil {
		return errors.New("driver app need settlement account service")
	}

	return nil
}
