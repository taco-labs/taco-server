package driver

import (
	"context"
	"errors"
	"fmt"
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

const (
	imageFileTpl = "driver-%s-%s"
)

func getImageFileName(driverId string, imageType enum.ImageType) string {
	return fmt.Sprintf(imageFileTpl, driverId, imageType)
}

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
	DriverCancelTaxiCallRequest(context.Context, string, string) error
}

type driverSettlementInterface interface {
	GetExpectedDriverSettlement(ctx context.Context, driverId string) (entity.DriverExpectedSettlement, error)
	ListDriverSettlementHistory(ctx context.Context, req request.ListDriverSettlementHistoryRequest) ([]entity.DriverSettlementHistory, time.Time, error)
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
		smsSender         service.SmsSenderService
		session           sessionServiceInterface
		push              pushServiceInterface
		taxiCall          driverTaxiCallInterface
		imageUrl          service.ImageUrlService
		settlementAccount service.SettlementAccountService // TODO (taekyeom) settlement app으로 이동
		driverSettlement  driverSettlementInterface
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
	var driverDto entity.DriverDto
	var driverSession entity.DriverSession

	err = d.RunWithNonRollbackError(ctx, value.ErrDriverNotFound, func(ctx context.Context, i bun.IDB) error {
		smsVerification, err = d.repository.smsVerification.FindById(ctx, i, req.StateKey)
		if err != nil {
			return fmt.Errorf("app.Driver.SmsSignin: error while find sms verification:\n %w", err)
		}

		if smsVerification.VerificationCode != req.VerificationCode {
			return fmt.Errorf("app.Driver.SmsSignin: invalid verification code:\n%w", value.ErrInvalidOperation)
		}
		driver, err := d.repository.driver.FindByUserUniqueKey(ctx, i, smsVerification.Phone)
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

		driverDto = driver

		return nil
	})

	if err != nil {
		return entity.Driver{}, "", err
	}

	downloadUrls, uploadUrls, err := d.driverImageUrls(ctx, driverDto.Id)
	if err != nil {
		return entity.Driver{}, "", err
	}

	driver := entity.Driver{
		DriverDto:    driverDto,
		DownloadUrls: downloadUrls,
		UploadUrls:   uploadUrls,
	}

	return driver, driverSession.Id, nil
}

func (d driverApp) Signup(ctx context.Context, req request.DriverSignupRequest) (entity.Driver, string, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	var newDriverDto entity.DriverDto
	var driverSession entity.DriverSession

	err := d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
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

		_, supportedRegion := value.SupportedServiceRegions[req.ServiceRegion]
		if !supportedRegion {
			return fmt.Errorf("app.Driver.Signup: unsupported service region: %w", value.ErrUnsupportedServiceRegion)
		}

		newDriverDto = entity.DriverDto{
			Id:                         utils.MustNewUUID(),
			DriverType:                 enum.DriverTypeFromString(req.DriverType),
			FirstName:                  req.FirstName,
			LastName:                   req.LastName,
			BirthDay:                   req.Birthday,
			Gender:                     req.Gender,
			Phone:                      req.Phone,
			AppOs:                      enum.OsTypeFromString(req.AppOs),
			AppVersion:                 req.AppVersion,
			UserUniqueKey:              req.Phone,
			DriverLicenseId:            req.DriverLicenseId,
			CompanyRegistrationNumber:  req.CompanyRegistrationNumber,
			CarNumber:                  req.CarNumber,
			ServiceRegion:              req.ServiceRegion,
			DriverLicenseImageUploaded: false,
			DriverProfileImageUploaded: false,
			OnDuty:                     false,
			Active:                     false,
			CreateTime:                 requestTime,
			UpdateTime:                 requestTime,
			DeleteTime:                 time.Time{},
		}

		if err := d.repository.driver.Create(ctx, i, newDriverDto); err != nil {
			return fmt.Errorf("app.Driver.Signup: error while create driver:%w", err)
		}

		// Create push token
		if _, err := d.service.push.CreatePushToken(ctx, request.CreatePushTokenRequest{
			PrincipalId: newDriverDto.Id,
			FcmToken:    req.AppFcmToken,
		}); err != nil {
			return fmt.Errorf("app.Driver.Signup: error while create push token: %w", err)
		}

		driverSession = entity.DriverSession{
			Id:         utils.MustNewUUID(),
			DriverId:   newDriverDto.Id,
			ExpireTime: requestTime.AddDate(0, 1, 0),
			Activated:  newDriverDto.Active,
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

	downloadUrls, uploadUrls, err := d.driverImageUrls(ctx, newDriverDto.Id)
	if err != nil {
		return entity.Driver{}, "", err
	}

	newDriver := entity.Driver{
		DriverDto:    newDriverDto,
		DownloadUrls: downloadUrls,
		UploadUrls:   uploadUrls,
	}

	return newDriver, driverSession.Id, nil
}

func (d driverApp) GetDriver(ctx context.Context, driverId string) (entity.Driver, error) {
	var driverDto entity.DriverDto
	var err error
	err = d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		driverDto, err = d.repository.driver.FindById(ctx, i, driverId)
		if err != nil {
			return fmt.Errorf("app.Driver.GetDriver: error while find driver by id:\n%w", err)
		}

		return nil
	})

	if err != nil {
		return entity.Driver{}, err
	}

	downloadUrls, uploadUrls, err := d.driverImageUrls(ctx, driverDto.Id)
	if err != nil {
		return entity.Driver{}, err
	}

	driver := entity.Driver{
		DriverDto:    driverDto,
		DownloadUrls: downloadUrls,
		UploadUrls:   uploadUrls,
	}

	return driver, nil
}

func (d driverApp) UpdateDriver(ctx context.Context, req request.DriverUpdateRequest) (entity.Driver, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	var driverDto entity.DriverDto
	var err error

	err = d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		driverDto, err = d.repository.driver.FindById(ctx, i, req.Id)
		if err != nil {
			return fmt.Errorf("app.driver.UpdateDriver: error while find driver by id:\n %w", err)
		}

		driverDto.AppOs = enum.OsTypeFromString(req.AppOs)
		driverDto.AppVersion = req.AppVersion
		driverDto.DriverLicenseImageUploaded = req.LicenseImageUploaded
		driverDto.DriverProfileImageUploaded = req.ProfileImageUploaded
		driverDto.CarNumber = req.CarNumber
		driverDto.UpdateTime = requestTime

		if err := d.repository.driver.Update(ctx, i, driverDto); err != nil {
			return fmt.Errorf("app.Driver.UpdateDriver: error while update driver: %w", err)
		}

		// Update push token
		if err := d.service.push.UpdatePushToken(ctx, request.UpdatePushTokenRequest{
			PrincipalId: driverDto.Id,
			FcmToken:    req.AppFcmToken,
		}); err != nil {
			return fmt.Errorf("app.Driver.UpdateDriver: error while update push token: %w", err)
		}

		return nil
	})

	if err != nil {
		return entity.Driver{}, err
	}

	downloadUrls, uploadUrls, err := d.driverImageUrls(ctx, driverDto.Id)
	if err != nil {
		return entity.Driver{}, err
	}

	driver := entity.Driver{
		DriverDto:    driverDto,
		DownloadUrls: downloadUrls,
		UploadUrls:   uploadUrls,
	}

	return driver, nil
}

func (d driverApp) GetDriverImageUrls(ctx context.Context, driverId string) (value.DriverImageUrls, value.DriverImageUrls, error) {
	downloadUrls, uploadUrls, err := d.driverImageUrls(ctx, driverId)
	if err != nil {
		return value.DriverImageUrls{}, value.DriverImageUrls{}, fmt.Errorf("app.driver.GetDriverImageUrls: error while get image urls: %w", err)
	}
	return downloadUrls, uploadUrls, nil
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
