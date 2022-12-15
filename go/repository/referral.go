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

type ReferralRepository interface {
	GetUserReferral(context.Context, bun.IDB, string) (entity.UserReferral, error)
	CreateUserReferral(context.Context, bun.IDB, entity.UserReferral) error
	UpdateUserReferral(context.Context, bun.IDB, entity.UserReferral) error
	DeleteUserReferral(context.Context, bun.IDB, entity.UserReferral) error

	GetDriverReferral(context.Context, bun.IDB, string) (entity.DriverReferral, error)
	CreateDriverReferral(context.Context, bun.IDB, entity.DriverReferral) error
	UpdateDriverReferral(context.Context, bun.IDB, entity.DriverReferral) error
	DeleteDriverReferral(context.Context, bun.IDB, entity.DriverReferral) error
}

type referralRepository struct{}

func (u referralRepository) GetUserReferral(ctx context.Context, db bun.IDB, userId string) (entity.UserReferral, error) {
	resp := entity.UserReferral{
		FromUserId: userId,
	}

	err := db.NewSelect().Model(&resp).WherePK().Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.UserReferral{}, value.ErrNotFound
	}
	if err != nil {
		return entity.UserReferral{}, fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (u referralRepository) CreateUserReferral(ctx context.Context, db bun.IDB, userReferral entity.UserReferral) error {
	res, err := db.NewInsert().Model(&userReferral).Exec(ctx)

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

func (u referralRepository) UpdateUserReferral(ctx context.Context, db bun.IDB, userReferral entity.UserReferral) error {
	res, err := db.NewUpdate().Model(&userReferral).WherePK().Exec(ctx)

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

func (u referralRepository) DeleteUserReferral(ctx context.Context, db bun.IDB, userReferral entity.UserReferral) error {
	res, err := db.NewDelete().Model(&userReferral).WherePK().Exec(ctx)

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

func (u referralRepository) GetDriverReferral(ctx context.Context, db bun.IDB, driverId string) (entity.DriverReferral, error) {
	resp := entity.DriverReferral{
		DriverId: driverId,
	}

	err := db.NewSelect().Model(&resp).WherePK().Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.DriverReferral{}, value.ErrNotFound
	}
	if err != nil {
		return entity.DriverReferral{}, fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (u referralRepository) CreateDriverReferral(ctx context.Context, db bun.IDB, driverReferral entity.DriverReferral) error {
	res, err := db.NewInsert().Model(&driverReferral).Exec(ctx)

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

func (u referralRepository) UpdateDriverReferral(ctx context.Context, db bun.IDB, driverReferral entity.DriverReferral) error {
	res, err := db.NewUpdate().Model(&driverReferral).WherePK().Exec(ctx)

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

func (u referralRepository) DeleteDriverReferral(ctx context.Context, db bun.IDB, driverReferral entity.DriverReferral) error {
	res, err := db.NewDelete().Model(&driverReferral).WherePK().Exec(ctx)

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

func NewReferralRepository() *referralRepository {
	return &referralRepository{}
}
