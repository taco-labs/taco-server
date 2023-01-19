package driver

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/domain/value/analytics"
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
	LatestDriverTaxiCallRequest(ctx context.Context, driverId string) (entity.DriverLatestTaxiCallRequest, error)
	ForceAcceptTaxiCallRequest(ctx context.Context, driverId, callRequestId string) (entity.DriverLatestTaxiCallRequest, error)
	AcceptTaxiCallRequest(ctx context.Context, driverId string, ticketId string) (entity.DriverLatestTaxiCallRequest, error)
	RejectTaxiCallRequest(ctx context.Context, driverId string, ticketId string) error
	DriverToArrival(ctx context.Context, driverId string, callRequestId string) error
	DoneTaxiCallRequest(ctx context.Context, driverId string, req request.DoneTaxiCallRequest) error
	DriverCancelTaxiCallRequest(context.Context, string, request.DriverCancelTaxiCallRequest) error
	DriverLatestTaxiCallTicket(ctx context.Context, driverId string) (entity.DriverLatestTaxiCallRequestTicket, error)
	AddDriverDenyTaxiCallTag(ctx context.Context, driverId string, tagId int) error
	DeleteDriverDenyTaxiCallTag(ctx context.Context, driverId string, tagId int) error
	ListDriverDenyTaxiCallTag(ctx context.Context, driverId string) ([]value.Tag, error)
}

type driverSettlementInterface interface {
	GetExpectedDriverSettlement(ctx context.Context, driverId string) (entity.DriverTotalSettlement, error)
	ListDriverSettlementHistory(ctx context.Context, req request.ListDriverSettlementHistoryRequest) ([]entity.DriverSettlementHistory, time.Time, error)
	RequestSettlementTransfer(ctx context.Context, settlementAccount entity.DriverSettlementAccount) (int, error)
	ReceiveDriverPromotionReward(ctx context.Context, driverId string, receiveTime time.Time) error
}

type userPaymentApp interface {
	CreateDriverReferral(ctx context.Context, fromDriverId, toDriverId string) error
}

type driverApp struct {
	app.Transactor
	repository struct {
		driver            repository.DriverRepository
		settlementAccount repository.DriverSettlementAccountRepository
		smsVerification   repository.SmsVerificationRepository
		event             repository.EventRepository
		analytics         repository.AnalyticsRepository
	}

	service struct {
		smsSender            service.SmsVerificationSenderService
		session              sessionServiceInterface
		push                 pushServiceInterface
		taxiCall             driverTaxiCallInterface
		imageUploadUrl       service.ImageUploadUrlService
		imageDownloadUrl     service.ImageDownloadUrlService
		settlementAccount    service.SettlementAccountService // TODO (taekyeom) settlement app으로 이동
		driverSettlement     driverSettlementInterface
		payment              userPaymentApp
		encryption           service.EncryptionService
		serviceRegionChecker service.ServiceRegionChecker
	}

	config struct {
		taxiCallEnabled bool
	}
}

// TODO (taekyeom) mock account logic seperation
func (d driverApp) SmsVerificationRequest(ctx context.Context, req request.SmsVerificationRequest) (entity.SmsVerification, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	var smsVerification entity.SmsVerification
	if value.IsMockPhoneNumber(req.Phone) {
		smsVerification = entity.NewMockSmsVerification(req.StateKey, requestTime, req.Phone)
	} else {
		code := d.service.smsSender.GenerateCode(6)
		smsVerification = entity.NewSmsVerification(req.StateKey, code, requestTime, req.Phone)
	}

	err := d.Run(ctx, func(ctx context.Context, db bun.IDB) error {
		if err := d.repository.smsVerification.Create(ctx, db, smsVerification); err != nil {
			return fmt.Errorf("app.Driver.SmsVerificationRequest: error while create sms verification:\n%w", err)
		}

		return nil
	})

	if !value.IsMockPhoneNumber(req.Phone) {
		if err := d.service.smsSender.SendSmsVerification(ctx, smsVerification); err != nil {
			return entity.SmsVerification{}, fmt.Errorf("app.Driver.SmsVerificationRequest: error while send sms message:\n%w", err)
		}
	}

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
	referralId := ""

	encryptedResidentRegistrationNumber, err := d.service.encryption.Encrypt(ctx, req.ResidentRegistrationNumber)
	if err != nil {
		return entity.Driver{}, "", fmt.Errorf("app.Driver.Signup: error while encrypt user resident registration number: %w", err)
	}

	err = d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		driver, err := d.repository.driver.FindByUserUniqueKey(ctx, i, req.Phone)
		if err != nil && !errors.Is(err, value.ErrDriverNotFound) {
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
			return fmt.Errorf("app.Driver.Signup: not verified phone:\n%w", value.ErrInvalidOperation)
		}

		driverId := utils.MustNewUUID()

		newDriverDto = entity.DriverDto{
			Id:                         driverId,
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
			ServiceRegion:              req.ServiceRegion,
			DriverLicenseImageUploaded: false,
			DriverProfileImageUploaded: false,
			OnDuty:                     false,
			Active:                     false,
			CreateTime:                 requestTime,
			UpdateTime:                 requestTime,
			DeleteTime:                 time.Time{},
		}

		if req.CarNumber != "" {
			newDriverCarProfile := entity.DriverCarProfile{
				Id:           utils.MustNewUUID(),
				DriverId:     driverId,
				TaxiCategory: req.TaxiCategory,
				CarNumber:    req.CarNumber,
				CarModel:     req.CarModel,
				CreateTime:   requestTime,
				UpdateTime:   requestTime,
			}
			newDriverDto.CarProfileId = newDriverCarProfile.Id
			newDriverDto.CarProfile = newDriverCarProfile
			if err := d.repository.driver.Create(ctx, i, newDriverDto); err != nil {
				return fmt.Errorf("app.Driver.Signup: error while create driver:%w", err)
			}

			if err := d.repository.driver.CreateDriverCarProfile(ctx, i, newDriverCarProfile); err != nil {
				return fmt.Errorf("app.Driver.Signup: error while create driver car profile:%w", err)
			}

		} else {
			if err := d.repository.driver.Create(ctx, i, newDriverDto); err != nil {
				return fmt.Errorf("app.Driver.Signup: error while create driver:%w", err)
			}
		}

		if err := d.repository.driver.CreateDriverRegistrationNumber(ctx, i, entity.DriverResidentRegistrationNumber{
			DriverId:                            newDriverDto.Id,
			EncryptedResidentRegistrationNumber: encryptedResidentRegistrationNumber,
		}); err != nil {
			return fmt.Errorf("app.Driver.Signup: error while create driver registration number:%w", err)
		}

		// Referral
		if req.ReferralCode != "" && !newDriverDto.MockAccount() {
			var referralCode string
			if strings.HasPrefix(req.ReferralCode, "010") {
				referralCode = req.ReferralCode
			} else {
				referralCode = fmt.Sprintf("010%s", req.ReferralCode)
			}

			referralDriver, err := d.repository.driver.FindByUserUniqueKey(ctx, i, referralCode)
			if errors.Is(err, value.ErrNotFound) || referralDriver.Id == newDriverDto.Id {
				return fmt.Errorf("app.Driver.Signup: driver not found: %w",
					value.NewTacoError(value.ERR_NOTFOUND_REFERRAL_CODE, "driver referral not found"))
			}
			if err != nil {
				return fmt.Errorf("app.Driver.Signup: error while get referral driver: %w", err)
			}

			if !referralDriver.MockAccount() {
				if err := d.service.payment.CreateDriverReferral(ctx, newDriverDto.Id, referralDriver.Id); err != nil {
					return fmt.Errorf("app.User.Signup: error while create user referral: %w", err)
				}
				referralId = referralDriver.Id
			}
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

		driverSignupAnalyticsEvent := entity.NewAnalytics(requestTime, analytics.DriverSignupPayload{
			DriverId:         newDriverDto.Id,
			ServiceRegion:    newDriverDto.ServiceRegion,
			ReferralDriverId: referralId,
		})

		if err := d.repository.analytics.Create(ctx, i, driverSignupAnalyticsEvent); err != nil {
			return fmt.Errorf("app.Driver.Signup: error while create analytics event: %w", err)
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

func (d driverApp) GetDriverByUserUniqueKey(ctx context.Context, uniqueKey string) (entity.DriverDto, error) {
	var driverDto entity.DriverDto
	var err error
	err = d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		driverDto, err = d.repository.driver.FindByUserUniqueKey(ctx, i, uniqueKey)
		if err != nil {
			return fmt.Errorf("app.Driver.GetDriverByUserUniqueKey: error while find driver by user unique key:\n%w", err)
		}

		return nil
	})

	if err != nil {
		return entity.DriverDto{}, err
	}

	return driverDto, nil
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
		driverDto.UpdateTime = requestTime

		if err := d.repository.driver.Update(ctx, i, driverDto); err != nil {
			return fmt.Errorf("app.Driver.UpdateDriver: error while update driver: %w", err)
		}

		// TODO (taekyeom) 하위 호환성을 위해 car number update로직을 남겨둠, 나중에 car profile로 migratrion한 후에 삭제 필요
		if driverDto.CarProfile.CarNumber != req.CarNumber {
			driverDto.CarProfile.CarNumber = req.CarNumber
			driverDto.CarProfile.UpdateTime = requestTime
			if err := d.repository.driver.UpdateDriverCarProfile(ctx, i, driverDto.CarProfile); err != nil {
				return fmt.Errorf("app.Driver.UpdateDriver: error while update driver car profile: %w", err)
			}
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

	if req.OnDuty && !d.config.taxiCallEnabled {
		return fmt.Errorf("app.Driver.UpdateOnDuty: temporarily unavailable: %w", value.ErrTemporarilyUnsupported)
	}

	return d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		driver, err := d.repository.driver.FindById(ctx, i, req.DriverId)
		if err != nil {
			return fmt.Errorf("app.Driver.UpdateOnDuty: error while find driver: %w", err)
		}

		if driver.OnDuty == req.OnDuty {
			return fmt.Errorf("app.Driver.UpdateOnDuty: requested on duty parameter is same as current state: %w", value.ErrInvalidOperation)
		}

		supportedRegion, err := d.service.serviceRegionChecker.CheckAvailableServiceRegion(ctx, driver.ServiceRegion)
		if err != nil {
			return fmt.Errorf("app.Driver.UpdateOnDuty: error while check service region: %w", err)
		}
		if !supportedRegion {
			return fmt.Errorf("app.Driver.UpdateOnDuty: unsupported service region: %w", value.ErrUnsupportedServiceRegion)
		}

		if driver.CarProfileId == "" {
			return fmt.Errorf("app.Driver.UpdateOnDuty: select car profile first: %w", value.ErrCarProfileNotSelected)
		}

		driver.OnDuty = req.OnDuty
		driverOnDutyPrevUpdateTime := driver.OnDutyUpdateTime
		driver.OnDutyUpdateTime = requestTime
		driver.UpdateTime = requestTime

		if err := d.repository.driver.Update(ctx, i, driver); err != nil {
			return fmt.Errorf("app.Driver.UpdateOnDuty: error while update driver:%w", err)
		}

		if req.OnDuty {
			if err := d.service.taxiCall.ActivateDriverContext(ctx, req.DriverId); err != nil {
				return err
			}
		} else {
			if err := d.service.taxiCall.DeactivateDriverContext(ctx, req.DriverId); err != nil {
				return err
			}
			if requestTime.Sub(driverOnDutyPrevUpdateTime) > time.Hour*4 {
				if err := d.service.driverSettlement.ReceiveDriverPromotionReward(ctx, driver.Id, driverOnDutyPrevUpdateTime); err != nil {
					return err
				}
			}
		}

		driverOnDutyAnalyticsEvent := entity.NewAnalytics(requestTime, analytics.DriverOnDutyPayload{
			DriverId: driver.Id,
			OnDuty:   driver.OnDuty,
			Duration: requestTime.Sub(driverOnDutyPrevUpdateTime),
		})
		if err := d.repository.analytics.Create(ctx, i, driverOnDutyAnalyticsEvent); err != nil {
			return fmt.Errorf("app.Driver.UpdateOnDuty: error while create analytics event: %w", err)
		}

		return nil
	})
}

func (d driverApp) UpdateDriverLocation(ctx context.Context, req request.DriverLocationUpdateRequest) error {
	err := d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		driver, err := d.repository.driver.FindById(ctx, i, req.DriverId)

		if err != nil {
			return fmt.Errorf("app.Driver.UpdateDriverLocation: error while find driver: %w", err)
		}
		if !driver.OnDuty {
			return fmt.Errorf("app.Driver.UpdateDriverLocation: driver is not on duty: %w", value.ErrInvalidOperation)
		}

		return nil
	})

	return d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		err = d.service.taxiCall.UpdateDriverLocation(ctx, req)

		if err != nil {
			return fmt.Errorf("app.Driver.UpdateDriverLocation: error while update driver location via taxicall app: %w", err)
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

		// TODO (taekyeom) use different error?
		if driver.Active {
			return fmt.Errorf("app.Driver.ActivateDriver: already activated: %w", value.ErrAlreadyExists)
		}

		driver.Active = true

		if err := d.repository.driver.Update(ctx, i, driver); err != nil {
			return fmt.Errorf("app.Driver.ActivateDriver: error while activate driver:%w", err)
		}

		if err := d.service.session.ActivateByDriverId(ctx, driverId); err != nil {
			return fmt.Errorf("app.Driver.ActivateDriver: error while activate driver session:%w", err)
		}

		driverActivatedPushNotification := NewDriverActivatedNotification(driver.Id)
		if err := d.repository.event.BatchCreate(ctx, i, []entity.Event{driverActivatedPushNotification}); err != nil {
			return fmt.Errorf("app.Driver.ActivateDriver: error while create driver activated push event: %w", err)
		}

		return nil
	})
}

func (d driverApp) ListNonActivatedDriver(ctx context.Context, req request.ListNonActivatedDriverRequest) ([]entity.DriverDto, string, error) {
	var driverDtos []entity.DriverDto
	var pageToken string
	err := d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		dtos, newPageToken, err := d.repository.driver.ListNotActivatedDriver(ctx, i, req.PageToken, req.Count)
		if err != nil {
			return fmt.Errorf("app.Driver.ListNonActivatedDriver: error while list non activated driver: %w", err)
		}

		driverDtos = dtos
		pageToken = newPageToken
		return nil
	})

	if err != nil {
		return []entity.DriverDto{}, "", err
	}

	return driverDtos, pageToken, nil
}

func (d driverApp) ListAvailableServiceRegion(ctx context.Context) ([]string, error) {
	serviceRegions, err := d.service.serviceRegionChecker.ListServiceRegion(ctx)

	if err != nil {
		return []string{}, fmt.Errorf("app.driverApp.ListAvailableServiceRegion: error while list available service region: %w", err)
	}
	return serviceRegions, nil
}
