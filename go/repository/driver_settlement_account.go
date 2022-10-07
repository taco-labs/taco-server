package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/uptrace/bun"
)

type DriverSettlementAccountRepository interface {
	GetByDriverId(context.Context, bun.IDB, string) (entity.DriverSettlementAccount, error)
	Create(context.Context, bun.IDB, entity.DriverSettlementAccount) error
	Update(context.Context, bun.IDB, entity.DriverSettlementAccount) error
}

type driverSettlementAccountRepository struct{}

func (d driverSettlementAccountRepository) GetByDriverId(ctx context.Context, db bun.IDB, driverId string) (entity.DriverSettlementAccount, error) {
	account := entity.DriverSettlementAccount{
		DriverId: driverId,
	}

	err := db.NewSelect().Model(&account).WherePK().Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return entity.DriverSettlementAccount{}, value.ErrNotFound
	}
	if err != nil {
		return entity.DriverSettlementAccount{}, fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return account, nil
}

func (d driverSettlementAccountRepository) Create(ctx context.Context, db bun.IDB, account entity.DriverSettlementAccount) error {
	res, err := db.NewInsert().Model(&account).Exec(ctx)

	if err != nil {
		return fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}
	if rowsAffected != 1 {
		return fmt.Errorf("%w: invalid rows affected %d", value.ErrDBInternal, rowsAffected)
	}

	return nil
}

func (d driverSettlementAccountRepository) Update(ctx context.Context, db bun.IDB, account entity.DriverSettlementAccount) error {
	res, err := db.NewUpdate().Model(&account).WherePK().Exec(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return value.ErrDriverNotFound
	}
	if err != nil {
		return fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}
	if rowsAffected != 1 {
		return fmt.Errorf("%w: invalid rows affected %d", value.ErrDBInternal, rowsAffected)
	}

	return nil
}

func NewDriverSettlementAccountRepository() driverSettlementAccountRepository {
	return driverSettlementAccountRepository{}
}
