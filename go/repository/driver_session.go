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

type DriverSessionRepository interface {
	GetById(context.Context, bun.IDB, string) (entity.DriverSession, error)
	ActivateByDriverId(context.Context, bun.IDB, string) error
	Create(context.Context, bun.IDB, entity.DriverSession) error
	DeleteByDriverId(context.Context, bun.IDB, string) error
}

type driverSessionRepository struct{}

func (d driverSessionRepository) GetById(ctx context.Context, db bun.IDB, sessionId string) (entity.DriverSession, error) {
	driverSession := entity.DriverSession{Id: sessionId}

	err := db.NewSelect().Model(&driverSession).WherePK().Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return entity.DriverSession{}, value.ErrNotFound
	}
	if err != nil {
		return entity.DriverSession{}, fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return driverSession, nil
}

func (d driverSessionRepository) ActivateByDriverId(ctx context.Context, db bun.IDB, driverId string) error {
	res, err := db.NewUpdate().
		Model(&entity.DriverSession{}).
		Set("activated = true").
		Where("driver_id = ?", driverId).
		Exec(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return value.ErrNotFound
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

func (d driverSessionRepository) Create(ctx context.Context, db bun.IDB, driverSession entity.DriverSession) error {
	res, err := db.NewInsert().Model(&driverSession).Exec(ctx)

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

func (d driverSessionRepository) DeleteByDriverId(ctx context.Context, db bun.IDB, driverId string) error {
	res, err := db.NewDelete().Model(&entity.DriverSession{}).Where("driver_id = ?", driverId).Exec(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return value.ErrNotFound
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

func NewDriverSessionRepository() driverSessionRepository {
	return driverSessionRepository{}
}
