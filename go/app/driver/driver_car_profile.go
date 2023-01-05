package driver

import (
	"context"
	"fmt"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/utils"
	"github.com/uptrace/bun"
)

func (d driverApp) GetCarProfile(ctx context.Context, profileId string) (entity.DriverCarProfile, error) {
	var carProfile entity.DriverCarProfile
	driverId := utils.GetDriverId(ctx)

	err := d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		profile, err := d.repository.driver.GetDriverCarProfile(ctx, i, profileId)
		if err != nil {
			return fmt.Errorf("app.driver.GetCarProfile: error while get car profile: %w", err)
		}

		if profile.DriverId != driverId {
			return fmt.Errorf("app.driver.GetCarProfile: invalid driver id for profile: %w", value.ErrUnAuthorized)
		}

		carProfile = profile

		return nil
	})

	if err != nil {
		return entity.DriverCarProfile{}, err
	}

	return carProfile, nil
}

func (d driverApp) ListCarProfile(ctx context.Context, driverId string) ([]entity.DriverCarProfile, error) {
	var carProfiles []entity.DriverCarProfile

	err := d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		profiles, err := d.repository.driver.ListDriverCarProfileByDriverId(ctx, i, driverId)
		if err != nil {
			return fmt.Errorf("app.driver.ListCarProfile: error while list car profile: %w", err)
		}

		carProfiles = profiles
		return nil
	})

	if err != nil {
		return []entity.DriverCarProfile{}, err
	}

	return carProfiles, nil
}

func (d driverApp) SelectCarProfile(ctx context.Context, driverId string, profileId string) (entity.DriverCarProfile, error) {
	var carProfile entity.DriverCarProfile

	err := d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		profile, err := d.repository.driver.GetDriverCarProfile(ctx, i, profileId)
		if err != nil {
			return fmt.Errorf("app.driver.SelectCarProfile: error while get driver car profile: %w", err)
		}

		if profile.DriverId != driverId {
			return fmt.Errorf("app.driver.SelectCarProfile: invalid driver id for profile: %w", value.ErrUnAuthorized)
		}

		driver, err := d.repository.driver.FindById(ctx, i, driverId)
		if err != nil {
			return fmt.Errorf("app.driver.SelectCarProfile: error while get driver: %w", err)
		}

		if driver.OnDuty {
			return fmt.Errorf("app.driver.SelectCarProfile: update profile is allowed when driver is not on duty: %w", value.ErrNotAllowed)
		}

		carProfile = profile
		carProfile.Selected = true
		if profile.Id == driver.CarProfileId {
			return nil
		}
		driver.CarProfileId = profile.Id
		if err := d.repository.driver.Update(ctx, i, driver); err != nil {
			return fmt.Errorf("app.driver.SelectCarProfile: error while update car profile: %w", err)
		}

		return nil
	})

	if err != nil {
		return entity.DriverCarProfile{}, err
	}

	return carProfile, nil
}

func (d driverApp) AddCarProfile(ctx context.Context, req request.AddCarProfileRequest) (entity.DriverCarProfile, error) {
	driverId := utils.GetDriverId(ctx)
	requestTime := utils.GetRequestTimeOrNow(ctx)

	carProfile := entity.DriverCarProfile{
		Id:         utils.MustNewUUID(),
		DriverId:   driverId,
		CarNumber:  req.CarNumber,
		CarType:    req.CarType,
		CreateTime: requestTime,
		UpdateTime: requestTime,
	}

	err := d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		if err := d.repository.driver.CreateDriverCarProfile(ctx, i, carProfile); err != nil {
			return fmt.Errorf("app.driver.AddCarProfile: error while create car profile: %w", err)
		}
		return nil
	})

	if err != nil {
		return entity.DriverCarProfile{}, nil
	}

	return carProfile, nil
}

func (d driverApp) UpdateCarProfile(ctx context.Context, req request.UpdateCarProfileRequest) (entity.DriverCarProfile, error) {
	driverId := utils.GetDriverId(ctx)
	requestTime := utils.GetRequestTimeOrNow(ctx)
	var carProfile entity.DriverCarProfile

	err := d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		profile, err := d.repository.driver.GetDriverCarProfile(ctx, i, req.ProfileId)
		if err != nil {
			return fmt.Errorf("app.driver.UpdateCarProfile: error while get driver car profile: %w", err)
		}

		if profile.DriverId != driverId {
			return fmt.Errorf("app.driver.UpdateCarProfile: invalid driver id for profile: %w", value.ErrUnAuthorized)
		}

		driver, err := d.repository.driver.FindById(ctx, i, driverId)
		if err != nil {
			return fmt.Errorf("app.driver.UpdateCarProfile: error while get driver: %w", err)
		}

		if driver.OnDuty {
			return fmt.Errorf("app.driver.UpdateCarProfile: update profile is allowed when driver is not on duty: %w", value.ErrNotAllowed)
		}

		profile.CarNumber = req.CarNumber
		profile.CarType = req.CarType
		profile.UpdateTime = requestTime

		if err := d.repository.driver.UpdateDriverCarProfile(ctx, i, profile); err != nil {
			return fmt.Errorf("app.driver.UpdateCarProfile: error while update car profile: %w", err)
		}

		carProfile = profile

		return nil
	})

	if err != nil {
		return entity.DriverCarProfile{}, nil
	}

	return carProfile, nil
}

func (d driverApp) DeleteCarProfile(ctx context.Context, profileId string) error {
	driverId := utils.GetDriverId(ctx)

	return d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		profile, err := d.repository.driver.GetDriverCarProfile(ctx, i, profileId)
		if err != nil {
			return fmt.Errorf("app.driver.DeleteeCarProfile: error while get car profile: %w", err)
		}
		if profile.DriverId != driverId {
			return fmt.Errorf("app.driver.DeleteCarProfile: invalid driver id for profile: %w", value.ErrUnAuthorized)
		}

		driver, err := d.repository.driver.FindById(ctx, i, driverId)
		if err != nil {
			return fmt.Errorf("app.driver.DeleteCarProfile: error while get driver: %w", err)
		}

		if driver.CarProfileId == profile.Id {
			return fmt.Errorf("app.driver.DeleteCarProfile: profile is selected: %w", value.ErrInUse)
		}

		if driver.OnDuty {
			return fmt.Errorf("app.driver.DeleteCarProfile: delete profile is allowed when driver is not on duty: %w", value.ErrNotAllowed)
		}

		profiles, err := d.repository.driver.ListDriverCarProfileByDriverId(ctx, i, driverId)
		if err != nil {
			return fmt.Errorf("app.driver.DeleteCarProfile: error while list car profile: %w", err)
		}
		if len(profiles) == 1 {
			return fmt.Errorf("app.driver.DeleteCarProfile: no car profiles remains..: %w", value.ErrInvalidOperation)
		}

		if err := d.repository.driver.DeleteDriverCarProfile(ctx, i, profile); err != nil {
			return fmt.Errorf("app.driver.DeleteCarProfile: error while delete car profile: %w", err)
		}

		return nil
	})
}
