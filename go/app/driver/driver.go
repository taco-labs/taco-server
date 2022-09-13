package driver

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ktk1012/taco/go/app"
	"github.com/ktk1012/taco/go/domain/entity"
	"github.com/ktk1012/taco/go/domain/request"
	"github.com/ktk1012/taco/go/domain/value"
	"github.com/ktk1012/taco/go/domain/value/enum"
	"github.com/ktk1012/taco/go/repository"
	"github.com/ktk1012/taco/go/service"
	"github.com/ktk1012/taco/go/utils"
)

type sessionInterface interface {
	RevokeByDriverId(context.Context, string) error
	Create(context.Context, entity.DriverSession) error
}

type driverApp struct {
	app.Transactor
	repository struct {
		driver            repository.DriverRepository
		driverLocation    repository.DriverLocationRepository
		settlementAccount repository.DriverSettlementAccountRepository
	}

	service struct {
		userIdentity service.UserIdentityService
		session      sessionInterface
	}

	actor struct {
	}
}

func (d driverApp) Signup(ctx context.Context, req request.DriverSignupRequest) (entity.Driver, string, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)
	userIdentity, err := d.service.userIdentity.GetUserIdentity(ctx, req.IamUid)
	if err != nil {
		return entity.Driver{}, "", err
	}

	ctx, err = d.Start(ctx)
	if err != nil {
		return entity.Driver{}, "", err
	}
	defer func() {
		err = d.Done(ctx, err)
	}()

	driver, err := d.repository.driver.FindByUserUniqueKey(ctx, userIdentity.UserUniqueKey)
	if !errors.Is(value.ErrNotFound, err) && err != nil {
		return entity.Driver{}, "", fmt.Errorf("app.Driver.Signup: error while find user by unique key:\n %v", err)
	}

	// Id              string          `bun:"id,pk"`
	// DriverType      enum.DriverType `bun:"driver_type"`
	// FirstName       string          `bun:"first_name"`
	// LastName        string          `bun:"last_name"`
	// BirthDay        string          `bun:"birthday"`
	// Phone           string          `bun:"phone"`
	// Gender          string          `bun:"gender"`
	// AppOs           enum.OsType     `bun:"app_os"`
	// AppVersion      string          `bun:"app_version"`
	// AppFcmToken     string          `bun:"app_fcm_token"`
	// UserUniqueKey   string          `bun:"user_unique_key"`
	// DriverLicenseId string          `bun:"driver_license_id"`
	// OnDuty          bool            `bun:"driver_on_duty"`
	// Active          bool            `bun:"active"`
	// CreateTime      time.Time       `bun:"create_time"`
	// UpdateTime      time.Time       `bun:"update_time"`
	// DeleteTime      time.Time       `bun:"delete_time"`
	if errors.Is(value.ErrNotFound, err) {
		newDriver := entity.Driver{
			Id:              utils.MustNewUUID(),
			DriverType:      enum.DriverTypeFromString(req.DriverType),
			FirstName:       req.FirstName,
			LastName:        req.LastName,
			BirthDay:        userIdentity.BirthDay,
			Phone:           userIdentity.Phone,
			Gender:          userIdentity.Gender,
			AppOs:           enum.OsTypeFromString(req.AppOs),
			AppVersion:      req.AppVersion,
			AppFcmToken:     req.AppFcmToken,
			UserUniqueKey:   userIdentity.UserUniqueKey,
			DriverLicenseId: req.DriverLicenseId,
			OnDuty:          false,
			Active:          false,
			CreateTime:      requestTime,
			UpdateTime:      requestTime,
			DeleteTime:      time.Time{},
		}
		if err = d.repository.driver.Create(ctx, newDriver); err != nil {
			return entity.Driver{}, "", fmt.Errorf("app.Driver.Signup: error while create new driver:\n%v", err)
		}

		driverSession := entity.DriverSession{
			Id:         utils.MustNewUUID(),
			DriverId:   newDriver.Id,
			Activated:  newDriver.Active,
			ExpireTime: requestTime.AddDate(0, 1, 0),
		}
		if err = d.service.session.Create(ctx, driverSession); err != nil {
			return entity.Driver{}, "", fmt.Errorf("app.Driver.Signup: error while create new session:\n %v", err)
		}

		return newDriver, driverSession.Id, nil
	} else {
		driver.AppOs = enum.OsTypeFromString(req.AppOs)
		driver.AppVersion = req.AppVersion
		driver.AppFcmToken = req.AppFcmToken
		driver.Phone = userIdentity.Phone

		driver.UpdateTime = requestTime

		if err = d.repository.driver.Update(ctx, driver); err != nil {
			return entity.Driver{}, "", fmt.Errorf("app.Driver.Signup: error while update user:\n%v", err)
		}

		driverSession := entity.DriverSession{
			Id:         utils.MustNewUUID(),
			DriverId:   driver.Id,
			Activated:  driver.Active,
			ExpireTime: requestTime.AddDate(0, 1, 0),
		}
		if err = d.service.session.RevokeByDriverId(ctx, driver.Id); err != nil {
			return entity.Driver{}, "", fmt.Errorf("app.Driver.Signup: error while revoke previous session:\n%v", err)
		}

		if err = d.service.session.Create(ctx, driverSession); err != nil {
			return entity.Driver{}, "", fmt.Errorf("app.Driver.Signup: error while create new session:\n%v", err)
		}

		return driver, driverSession.Id, nil
	}
}

func (d driverApp) GetDriver(ctx context.Context, driverId string) (entity.Driver, error) {
	ctx, err := d.Start(ctx)
	if err != nil {
		return entity.Driver{}, err
	}
	defer func() {
		err = d.Done(ctx, err)
	}()

	driver, err := d.repository.driver.FindById(ctx, driverId)
	if err != nil {
		return entity.Driver{}, fmt.Errorf("app.Driver.GetDriver: error while find driver by id:\n%v", err)
	}

	return driver, nil
}

func (d driverApp) UpdateOnDuty(ctx context.Context, req request.DriverOnDutyUpdateRequest) error {
	ctx, err := d.Start(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = d.Done(ctx, err)
	}()

	driver, err := d.repository.driver.FindById(ctx, req.DriverId)
	if err != nil {
		return fmt.Errorf("app.Driver.UpdateOnDuty: error while find driver by id:\n%v", err)
	}

	driver.OnDuty = req.OnDuty

	if err = d.repository.driver.Update(ctx, driver); err != nil {
		return fmt.Errorf("app.Driver.UpdateOnDuty: error while update user:\n%v", err)
	}

	return nil
}

func (d driverApp) UpdateDriverLocation(ctx context.Context, req request.DriverLocationUpdateRequest) error {
	ctx, err := d.Start(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = d.Done(ctx, err)
	}()

	_, err = d.repository.driver.FindById(ctx, req.DriverId)
	if err != nil {
		return fmt.Errorf("app.Driver.UpdateDriverLocation: error while find driver by id:\n%v", err)
	}

	driverLocation := entity.NewDriverLocation(req.DriverId, req.Latitude, req.Longitude)

	if err = d.repository.driverLocation.Upsert(ctx, driverLocation); err != nil {
		return fmt.Errorf("app.Driver.UpdateDriverLocation: error while update driver location:\n%v", err)
	}

	return nil
}

func NewDriverApp(opts ...driverOption) (driverApp, error) {
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

	if d.repository.settlementAccount == nil {
		return errors.New("driver app need settlement account repository")
	}

	if d.service.userIdentity == nil {
		return errors.New("driver app need user identity service")
	}

	if d.service.session == nil {
		return errors.New("driver app need driver session service")
	}

	return nil
}
