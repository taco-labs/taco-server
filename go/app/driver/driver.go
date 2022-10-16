package driver

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	// "time"

	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/domain/value/enum"
	"github.com/taco-labs/taco/go/service"
	"github.com/taco-labs/taco/go/utils"
	"github.com/uptrace/bun"

	"github.com/taco-labs/taco/go/repository"
)

type sessionInterface interface {
	RevokeByDriverId(context.Context, string) error
	Create(context.Context, entity.DriverSession) error
	ActivateByDriverId(context.Context, string) error
}

type pushInterface interface {
	SendTaxiCallRequestAccept(context.Context, entity.TaxiCallRequest, entity.DriverTaxiCallContext) error
	SendTaxiCallRequestDone(context.Context, entity.TaxiCallRequest) error
}

type driverApp struct {
	app.Transactor
	repository struct {
		driver            repository.DriverRepository
		driverLocation    repository.DriverLocationRepository
		settlementAccount repository.DriverSettlementAccountRepository
		smsVerification   repository.SmsVerificationRepository
		taxiCallRequest   repository.TaxiCallRepository
		event             repository.EventRepository
	}

	service struct {
		smsSender  service.SmsSenderService
		session    sessionInterface
		fileUpload service.FileUploadService
	}

	actor struct {
	}
}

func (d driverApp) SmsVerificationRequest(ctx context.Context, req request.SmsVerificationRequest) (entity.SmsVerification, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	smsVerification := entity.NewSmsVerification(req.StateKey, requestTime, req.Phone)

	err := d.Run(ctx, func(ctx context.Context, db bun.IDB) error {
		if err := d.repository.smsVerification.Create(ctx, db, smsVerification); err != nil {
			return fmt.Errorf("app.Driver.SmsVerificationRequest: error while create sms verification:\n%w", err)
		}

		if err := d.service.smsSender.SendSms(ctx, req.Phone, smsVerification.VerificationCode); err != nil {
			return fmt.Errorf("app.Driver.SmsVerificationRequest: error while send sms message:\n%w", err)
		}
		return nil
	})

	if err != nil {
		return entity.SmsVerification{}, err
	}

	return smsVerification, nil
}

func (d driverApp) SmsSignin(ctx context.Context, req request.SmsSigninRequest) (entity.Driver, string, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	var smsVerification entity.SmsVerification
	var err error
	var driver entity.Driver
	var driverSession entity.DriverSession

	err = d.RunWithNonRollbackError(ctx, value.ErrDriverNotFound, func(ctx context.Context, i bun.IDB) error {
		smsVerification, err = d.repository.smsVerification.FindById(ctx, i, req.StateKey)
		if err != nil {
			return fmt.Errorf("app.Driver.SmsSignin: error while find sms verification:\n %w", err)
		}

		if smsVerification.VerificationCode != req.VerificationCode {
			return fmt.Errorf("app.Driver.SmsSignin: invalid verification code:\n%w", value.ErrInvalidOperation)
		}
		driver, err = d.repository.driver.FindByUserUniqueKey(ctx, i, smsVerification.Phone)
		if errors.Is(value.ErrDriverNotFound, err) {
			smsVerification.Verified = true
			if err := d.repository.smsVerification.Update(ctx, i, smsVerification); err != nil {
				return fmt.Errorf("app.Driver.SmsSingin: error while update sms verification:\n%w", err)
			}
			return err
		}
		if err != nil {
			return fmt.Errorf("app.Driver.SmsSignin: error while find driver by unique key\n%w", err)
		}

		// create new session
		driverSession = entity.DriverSession{
			Id:         utils.MustNewUUID(),
			DriverId:   driver.Id,
			ExpireTime: requestTime.AddDate(0, 1, 0), // TODO(taekyeom) Configurable expire time
			Activated:  driver.Active,
		}

		if err = d.repository.smsVerification.Delete(ctx, i, smsVerification); err != nil {
			return fmt.Errorf("app.Driver.SmsSignin: error while delete sms verification:\n%w", err)
		}

		// revoke session
		if err = d.service.session.RevokeByDriverId(ctx, driver.Id); err != nil {
			return fmt.Errorf("app.Driver.SmsSignin: error while revoke previous session:\n %w", err)
		}

		if err = d.service.session.Create(ctx, driverSession); err != nil {
			return fmt.Errorf("app.Driver.SmsSignin: error while create new session:\n %w", err)
		}

		return nil
	})

	if err != nil {
		return entity.Driver{}, "", err
	}

	return driver, driverSession.Id, nil
}

func (d driverApp) Signup(ctx context.Context, req request.DriverSignupRequest) (entity.Driver, string, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	var newDriver entity.Driver
	var driverSession entity.DriverSession

	// TODO (taekyeom) Handle delete when signup failed
	imageUrl, err := d.service.fileUpload.Upload(ctx, os.File{})
	if err != nil {
		return entity.Driver{}, "", fmt.Errorf("app.Driver.Signup: error while upload driver url: %w", err)
	}

	err = d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		driver, err := d.repository.driver.FindByUserUniqueKey(ctx, i, req.Phone)
		if !errors.Is(err, value.ErrDriverNotFound) {
			return fmt.Errorf("app.Driver.Signup: error while find driver by unique key:%w", err)
		}

		if driver.Id != "" {
			return fmt.Errorf("app.Driver.Signup: user already exists: %w", value.ErrAlreadyExists)
		}

		smsVerification, err := d.repository.smsVerification.FindById(ctx, i, req.SmsVerificationStateKey)
		if err != nil {
			return fmt.Errorf("app.Driver.Signup: failed to find sms verification:\n%w", err)
		}

		if !smsVerification.Verified {
			return fmt.Errorf("app.Driver.Signup: not verified phone:\n%w", value.ErrUnAuthorized)
		}

		newDriver = entity.Driver{
			Id:                    utils.MustNewUUID(),
			DriverType:            enum.DriverTypeFromString(req.DriverType),
			FirstName:             req.FirstName,
			LastName:              req.LastName,
			BirthDay:              req.Birthday,
			Gender:                req.Gender,
			Phone:                 req.Phone,
			AppOs:                 enum.OsTypeFromString(req.AppOs),
			AppVersion:            req.AppVersion,
			AppFcmToken:           req.AppFcmToken,
			UserUniqueKey:         req.Phone,
			DriverLicenseId:       req.DriverLicenseId,
			DriverLicenseImageUrl: imageUrl,
			OnDuty:                false,
			Active:                false,
			CreateTime:            requestTime,
			UpdateTime:            requestTime,
			DeleteTime:            time.Time{},
		}

		if err := d.repository.driver.Create(ctx, i, newDriver); err != nil {
			return fmt.Errorf("app.Driver.Signup: error while create driver:%w", err)
		}

		driverSession = entity.DriverSession{
			Id:         utils.MustNewUUID(),
			DriverId:   newDriver.Id,
			ExpireTime: requestTime.AddDate(0, 1, 0),
			Activated:  newDriver.Active,
		}
		if err := d.service.session.Create(ctx, driverSession); err != nil {
			return fmt.Errorf("app.Driver.Signup: error while create new session:%w", err)
		}

		if err := d.repository.smsVerification.Delete(ctx, i, smsVerification); err != nil {
			return fmt.Errorf("app.Driver.Signup: error while delete sms verification:%w", err)
		}

		return nil
	})

	if err != nil {
		return entity.Driver{}, "", err
	}

	return newDriver, driverSession.Id, nil
}

func (d driverApp) GetDriver(ctx context.Context, driverId string) (entity.Driver, error) {
	var driver entity.Driver
	var err error
	err = d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		driver, err = d.repository.driver.FindById(ctx, i, driverId)
		if err != nil {
			return fmt.Errorf("app.Driver.GetDriver: error while find driver by id:\n%w", err)
		}

		return nil
	})

	if err != nil {
		return entity.Driver{}, err
	}

	return driver, nil
}

func (d driverApp) UpdateDriver(ctx context.Context, req request.DriverUpdateRequest) (entity.Driver, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	var driver entity.Driver
	var err error

	err = d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		driver, err = d.repository.driver.FindById(ctx, i, req.Id)
		if err != nil {
			return fmt.Errorf("app.driver.UpdateDriver: error while find driver by id:\n %w", err)
		}

		driver.AppOs = enum.OsTypeFromString(req.AppOs)
		driver.AppVersion = req.AppVersion
		driver.AppFcmToken = req.AppFcmToken
		driver.UpdateTime = requestTime

		if err := d.repository.driver.Update(ctx, i, driver); err != nil {
			return fmt.Errorf("app.Driver.UpdateDriver: error while update driver: %w", err)
		}

		return nil
	})

	if err != nil {
		return entity.Driver{}, err
	}

	return driver, nil
}

func (d driverApp) UpdateOnDuty(ctx context.Context, req request.DriverOnDutyUpdateRequest) error {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	return d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		driver, err := d.repository.driver.FindById(ctx, i, req.DriverId)
		if err != nil {
			return fmt.Errorf("app.Driver.UpdateOnDuty: error while find driver: %w", err)
		}

		if driver.OnDuty == req.OnDuty {
			return nil
		}

		driver.OnDuty = req.OnDuty
		driver.UpdateTime = requestTime

		if !driver.OnDuty {
			lastTaxiCallRequest, err := d.repository.taxiCallRequest.GetLatestByDriverId(ctx, i, driver.Id)
			if err != nil && !errors.Is(err, value.ErrNotFound) {
				return fmt.Errorf("app.Driver.UpdateOnDuty: error while get last call request: %w", err)
			}
			if lastTaxiCallRequest.CurrentState.Active() {
				return fmt.Errorf("app.Driver.UpdateOnDuty: active taxi call request exists: %w", value.ErrActiveTaxiCallRequestExists)
			}
		}

		taxiCallContext, err := d.repository.taxiCallRequest.GetDriverTaxiCallContext(ctx, i, driver.Id)
		if err != nil && !errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("app.Driver.UpdateOnDuty: error while get last call request: %w", err)
		}
		if errors.Is(err, value.ErrNotFound) {
			taxiCallContext = entity.NewEmptyDriverTaxiCallContext(driver.Id, driver.OnDuty, requestTime)
		}
		if err := d.repository.taxiCallRequest.UpsertDriverTaxiCallContext(ctx, i, taxiCallContext); err != nil {
			return fmt.Errorf("app.Driver.UpdateOnDuty: error while upsert driver taxi call context: %w", err)
		}

		if err := d.repository.driver.Update(ctx, i, driver); err != nil {
			return fmt.Errorf("app.Driver.UpdateOnDuty: error while update driver:%w", err)
		}

		return nil
	})
}

func (d driverApp) UpdateDriverLocation(ctx context.Context, req request.DriverLocationUpdateRequest) error {
	return d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		driver, err := d.repository.driver.FindById(ctx, i, req.DriverId)

		if err != nil {
			return fmt.Errorf("app.Driver.UpdateDriverLocation: error while find driver: %w", err)
		}
		if driver.OnDuty {
			return fmt.Errorf("app.Driver.UpdateDriverLocation: driver is not on duty: %w", value.ErrInvalidOperation)
		}

		driverLocationDto := entity.DriverLocation{
			DriverId: req.DriverId,
			Location: value.Point{
				Latitude:  req.Latitude,
				Longitude: req.Longitude,
			},
		}

		if err = d.repository.driverLocation.Upsert(ctx, i, driverLocationDto); err != nil {
			return fmt.Errorf("app.Driver.UpdateDriverLocation: error while update driver location:\n%w", err)
		}

		return nil
	})
}

func (d driverApp) DeleteDriver(ctx context.Context, driverId string) error {
	return d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		driver, err := d.repository.driver.FindById(ctx, i, driverId)
		if err != nil {
			return fmt.Errorf("app.Driver.DeleteDriver: error while find driver by id:%w", err)
		}

		if err := d.repository.driver.Delete(ctx, i, driver); err != nil {
			return fmt.Errorf("app.Driver.DeleteDriver: error while delete driver: %w", err)
		}

		return nil
	})
}

func (d driverApp) ActivateDriver(ctx context.Context, driverId string) error {
	return d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		driver, err := d.repository.driver.FindById(ctx, i, driverId)
		if err != nil {
			return fmt.Errorf("app.Driver.ActivateDriver: error while find driver by id:%w", err)
		}
		driver.Active = true

		if err := d.repository.driver.Update(ctx, i, driver); err != nil {
			return fmt.Errorf("app.Driver.ActivateDriver: error while activate driver:%w", err)
		}

		if err := d.service.session.ActivateByDriverId(ctx, driverId); err != nil {
			return fmt.Errorf("app.Driver.ActivateDriver: error while activate driver session:%w", err)
		}

		return nil
	})
}

func NewDriverApp(opts ...driverAppOption) (driverApp, error) {
	da := driverApp{}

	for _, opt := range opts {
		opt(&da)
	}

	return da, da.validateApp()
}

func (d driverApp) validateApp() error {
	if d.Transactor == nil {
		return errors.New("driver app need transactor")
	}

	if d.repository.driver == nil {
		return errors.New("driver app need driver repository")
	}

	if d.repository.driverLocation == nil {
		return errors.New("driver app need driver location repostiroy")
	}

	if d.repository.smsVerification == nil {
		return errors.New("driver app need sms verification repository")
	}

	if d.repository.settlementAccount == nil {
		return errors.New("driver app need settlement account repository")
	}

	if d.repository.taxiCallRequest == nil {
		return errors.New("driver app need taxi call request repository")
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

	if d.service.fileUpload == nil {
		return errors.New("driver app need file upload service")
	}

	return nil
}
