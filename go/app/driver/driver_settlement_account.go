package driver

import (
	"context"
	"fmt"
	"time"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/utils"
	"github.com/uptrace/bun"
)

func (d driverApp) GetDriverSettlementAccount(ctx context.Context, driverId string) (entity.DriverSettlementAccount, error) {
	var driverSettlementAccount entity.DriverSettlementAccount
	err := d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		settlemtnAccount, err := d.repository.settlementAccount.GetByDriverId(ctx, i, driverId)
		if err != nil {
			return fmt.Errorf("app.Driver.GetDriverSettlementAccount: error while get settlemtn account:%w", err)
		}
		driverSettlementAccount = settlemtnAccount
		return nil
	})

	if err != nil {
		return entity.DriverSettlementAccount{}, err
	}

	return driverSettlementAccount, nil
}

func (d driverApp) RegisterDriverSettlementAccount(ctx context.Context,
	req request.DriverSettlementAccountRegisterRequest) (entity.DriverSettlementAccount, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	var driver entity.Driver

	err := d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		dr, err := d.repository.driver.FindById(ctx, i, req.DriverId)
		if err != nil {
			return fmt.Errorf("app.Driver.RegisterDriverSettlementAccount: error while find driver by id:%w", err)
		}
		driver = entity.Driver{
			DriverDto: dr,
		}
		return nil
	})

	if err != nil {
		return entity.DriverSettlementAccount{}, err
	}

	driverSettlementAccount := entity.DriverSettlementAccount{
		DriverId:      req.DriverId,
		Bank:          req.Bank,
		AccountNumber: req.AccountNumber,
		CreateTime:    requestTime,
		UpdateTime:    requestTime,
	}

	authorizedAccount, err := d.service.settlementAccount.AuthorizeSettlementAccount(
		ctx, driver, driverSettlementAccount,
	)
	if err != nil {
		return entity.DriverSettlementAccount{}, fmt.Errorf("app.Driver.RegisterDriverSettlementAccount: error while authorize settlement account: %w", err)
	}

	if !authorizedAccount {
		// TODO (taekyeom) 별도 error code 부여 필요
		return entity.DriverSettlementAccount{}, fmt.Errorf("app.Driver.RegisterDriverSettlementAccount: bank account name is different: %w", value.ErrInvalidOperation)
	}

	err = d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		if err := d.repository.settlementAccount.Create(ctx, i, driverSettlementAccount); err != nil {
			return fmt.Errorf("app.Driver.RegisterDriverSettlementAccount: error while create settlement account: %w", err)
		}

		return nil
	})

	if err != nil {
		return entity.DriverSettlementAccount{}, err
	}

	return driverSettlementAccount, nil
}

func (d driverApp) UpdateDriverSettlementAccount(ctx context.Context,
	req request.DriverSettlementAccountUpdateRequest) (entity.DriverSettlementAccount, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	var driverSettlementAccount entity.DriverSettlementAccount
	var driver entity.Driver
	err := d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		dr, err := d.repository.driver.FindById(ctx, i, req.DriverId)
		if err != nil {
			return fmt.Errorf("app.Driver.UpdateDriverSettlementAccount: error while find driver by id:%w", err)
		}
		driver = entity.Driver{
			DriverDto: dr,
		}

		account, err := d.repository.settlementAccount.GetByDriverId(ctx, i, req.DriverId)
		if err != nil {
			return fmt.Errorf("app.Driver.UpdateDriverSettlementAccount: error while get settlement account:%w", err)
		}

		driverSettlementAccount = account
		return nil
	})

	if err != nil {
		return entity.DriverSettlementAccount{}, err
	}

	authorizedAccount, err := d.service.settlementAccount.AuthorizeSettlementAccount(
		ctx, driver, driverSettlementAccount,
	)
	if err != nil {
		return entity.DriverSettlementAccount{}, fmt.Errorf("app.Driver.UpdateDriverSettlementAccount: error while authorize settlement account: %w", err)
	}

	if !authorizedAccount {
		// TODO (taekyeom) 별도 error code 부여 필요
		return entity.DriverSettlementAccount{}, fmt.Errorf("app.Driver.UpdateDriverSettlementAccount: bank account name is different: %w", value.ErrUnAuthorized)
	}

	err = d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		driverSettlementAccount.Bank = req.Bank
		driverSettlementAccount.AccountNumber = req.AccountNumber
		driverSettlementAccount.UpdateTime = requestTime

		if err := d.repository.settlementAccount.Update(ctx, i, driverSettlementAccount); err != nil {
			return fmt.Errorf("app.Driver.UpdateDriverSettlementAccount: error while update settlement account:%w", err)
		}
		return nil
	})

	if err != nil {
		return entity.DriverSettlementAccount{}, err
	}

	return driverSettlementAccount, nil
}

func (d driverApp) GetExpectedDriverSettlement(ctx context.Context, driverId string) (entity.DriverExpectedSettlement, error) {
	return d.service.driverSettlement.GetExpectedDriverSettlement(ctx, driverId)
}

func (d driverApp) ListDriverSettlementHistory(ctx context.Context, req request.ListDriverSettlementHistoryRequest) ([]entity.DriverSettlementHistory, time.Time, error) {
	return d.service.driverSettlement.ListDriverSettlementHistory(ctx, req)
}
