package driver

import (
	"context"
	"fmt"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/request"
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

	var driverSettlementAccount entity.DriverSettlementAccount
	err := d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		_, err := d.repository.driver.FindById(ctx, i, req.DriverId)
		if err != nil {
			return fmt.Errorf("app.Driver.RegisterDriverSettlementAccount: error while find driver by id:%w", err)
		}

		driverSettlementAccount = entity.DriverSettlementAccount{
			DriverId:      req.DriverId,
			Bank:          req.Bank,
			AccountNumber: req.AccountNumber,
			CreateTime:    requestTime,
			UpdateTime:    requestTime,
		}

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
	err := d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		account, err := d.repository.settlementAccount.GetByDriverId(ctx, i, req.DriverId)
		if err != nil {
			return fmt.Errorf("app.Driver.UpdateDriverSettlemtnAccount: error while get settlement account:%w", err)
		}

		account.Bank = req.Bank
		account.AccountNumber = req.AccountNumber
		account.UpdateTime = requestTime

		if err := d.repository.settlementAccount.Update(ctx, i, account); err != nil {
			return fmt.Errorf("app.Driver.UpdateDriverSettlemtnAccount: error while update settlement account:%w", err)
		}
		driverSettlementAccount = account
		return nil
	})

	if err != nil {
		return entity.DriverSettlementAccount{}, err
	}

	return driverSettlementAccount, nil
}
