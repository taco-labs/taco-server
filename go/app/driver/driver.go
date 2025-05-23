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

type sessionServiceInterface interface {
	RevokeByDriverId(context.Context, string) error
	Create(context.Context, entity.DriverSession) error
	ActivateByDriverId(context.Context, string) error
}

type pushServiceInterface interface {
	CreatePushToken(context.Context, request.CreatePushTokenRequest) (entity.PushToken, error)
	UpdatePushToken(context.Context, request.UpdatePushTokenRequest) error
	DeletePushToken(context.Context, string) error
}

type driverTaxiCallInterface interface {
	ActivateDriverContext(ctx context.Context, driverId string) error
	DeactivateDriverContext(ctx context.Context, driverId string) error
	UpdateDriverLocation(ctx context.Context, req request.DriverLocationUpdateRequest) error
	ListDriverTaxiCallRequest(ctx context.Context, req request.ListDriverTaxiCallRequest) ([]entity.TaxiCallRequest, string, error)
	LatestDriverTaxiCallRequest(ctx context.Context, driverId string) (entity.TaxiCallRequest, error)
	ForceAcceptTaxiCallRequest(ctx context.Context, driverId, callRequestId string) error
	AcceptTaxiCallRequest(ctx context.Context, driverId string, ticketId string) error
	RejectTaxiCallRequest(ctx context.Context, driverId string, ticketId string) error
	DriverToArrival(ctx context.Context, driverId string, callRequestId string) error
	DoneTaxiCallRequest(ctx context.Context, driverId string, req request.DoneTaxiCallRequest) error
}

type driverApp struct {
	app.Transactor
	repository struct {
		driver            repository.DriverRepository
		settlementAccount repository.DriverSettlementAccountRepository
		smsVerification   repository.SmsVerificationRepository
		event             repository.EventRepository
	}

	service struct {
		smsSender  service.SmsSenderService
		session    sessionServiceInterface
		fileUpload service.FileUploadService
		push       pushServiceInterface
		taxiCall   driverTaxiCallInterface
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

		// Create push token
		if _, err := d.service.push.CreatePushToken(ctx, request.CreatePushTokenRequest{
			PrincipalId: newDriver.Id,
			FcmToken:    req.AppFcmToken,
		}); err != nil {
			return fmt.Errorf("app.Driver.Signup: error while create push token: %w", err)
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

		// Update push token
		if err := d.service.push.UpdatePushToken(ctx, request.UpdatePushTokenRequest{
			PrincipalId: driver.Id,
			FcmToken:    req.AppFcmToken,
		}); err != nil {
			return fmt.Errorf("app.Driver.UpdateDriver: error while update push token: %w", err)
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

		if err := d.repository.driver.Update(ctx, i, driver); err != nil {
			return fmt.Errorf("app.Driver.UpdateOnDuty: error while update driver:%w", err)
		}

		if req.OnDuty {
			return d.service.taxiCall.ActivateDriverContext(ctx, req.DriverId)
		} else {
			return d.service.taxiCall.DeactivateDriverContext(ctx, req.DriverId)
		}
	})
}

func (d driverApp) UpdateDriverLocation(ctx context.Context, req request.DriverLocationUpdateRequest) error {
	return d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		driver, err := d.repository.driver.FindById(ctx, i, req.DriverId)

		if err != nil {
			return fmt.Errorf("app.Driver.UpdateDriverLocation: error while find driver: %w", err)
		}
		if !driver.OnDuty {
			return fmt.Errorf("app.Driver.UpdateDriverLocation: driver is not on duty: %w", value.ErrInvalidOperation)
		}

		return d.service.taxiCall.UpdateDriverLocation(ctx, req)
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

		// Delete push token
		if err := d.service.push.DeletePushToken(ctx, driverId); err != nil {
			return fmt.Errorf("app.Driver.DeleteDriver: error while delete push token: %w", err)
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
