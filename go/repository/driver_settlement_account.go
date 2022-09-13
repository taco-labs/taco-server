package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ktk1012/taco/go/domain/entity"
	"github.com/ktk1012/taco/go/domain/value"
)

type DriverSettlementAccountRepository interface {
	GetByDriverId(context.Context, string) (entity.DriverSettlementAccount, error)
	Create(context.Context, entity.DriverSettlementAccount) error
	Update(context.Context, entity.DriverSettlementAccount) error
}

type driverSettlementAccountRepository struct{}

func (d driverSettlementAccountRepository) GetByDriverId(ctx context.Context, driverId string) (entity.DriverSettlementAccount, error) {
	db := GetQueryContext(ctx)

	account := entity.DriverSettlementAccount{
		DriverId: driverId,
	}

	err := db.NewSelect().Model(&account).WherePK().Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return entity.DriverSettlementAccount{}, value.ErrDriverNotFound
	}
	if err != nil {
		return entity.DriverSettlementAccount{}, fmt.Errorf("%v: %v", value.ErrDBInternal, err)
	}

	return account, nil
}

func (d driverSettlementAccountRepository) Create(ctx context.Context, account entity.DriverSettlementAccount) error {
	db := GetQueryContext(ctx)

	res, err := db.NewInsert().Model(&account).Exec(ctx)

	if err != nil {
		return fmt.Errorf("%v: %v", value.ErrDBInternal, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%v: %v", value.ErrDBInternal, err)
	}
	if rowsAffected != 1 {
		return fmt.Errorf("%v: invalid rows affected %d", value.ErrDBInternal, rowsAffected)
	}

	return nil
}

func (d driverSettlementAccountRepository) Update(ctx context.Context, account entity.DriverSettlementAccount) error {
	db := GetQueryContext(ctx)

	res, err := db.NewUpdate().Model(&account).WherePK().Exec(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return value.ErrDriverNotFound
	}
	if err != nil {
		return fmt.Errorf("%v: %v", value.ErrDBInternal, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%v: %v", value.ErrDBInternal, err)
	}
	if rowsAffected != 1 {
		return fmt.Errorf("%v: invalid rows affected %d", value.ErrDBInternal, rowsAffected)
	}

	return nil
}

func NewDriverSettlementAccountRepository() driverSettlementAccountRepository {
	return driverSettlementAccountRepository{}
}
