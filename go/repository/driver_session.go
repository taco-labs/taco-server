package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ktk1012/taco/go/domain/entity"
	"github.com/ktk1012/taco/go/domain/value"
)

type DriverSessionRepository interface {
	GetById(context.Context, string) (entity.DriverSession, error)
	ActivateByDriverId(context.Context, string) error
	Create(context.Context, entity.DriverSession) error
	DeleteByDriverId(context.Context, string) error
}

type driverSessionRepository struct{}

func (d driverSessionRepository) GetById(ctx context.Context, sessionId string) (entity.DriverSession, error) {
	db := GetQueryContext(ctx)

	driverSession := entity.DriverSession{Id: sessionId}

	err := db.NewSelect().Model(&driverSession).WherePK().Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return entity.DriverSession{}, value.ErrNotFound
	}
	if err != nil {
		return entity.DriverSession{}, fmt.Errorf("%v: %v", value.ErrDBInternal, err)
	}

	return driverSession, nil
}

func (d driverSessionRepository) ActivateByDriverId(ctx context.Context, driverId string) error {
	db := GetQueryContext(ctx)

	res, err := db.NewUpdate().
		Model(&entity.DriverSession{}).
		Set("activated = true").
		Where("driver_id = ?", driverId).
		Exec(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return value.ErrNotFound
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

func (d driverSessionRepository) Create(ctx context.Context, driverSession entity.DriverSession) error {
	db := GetQueryContext(ctx)

	res, err := db.NewInsert().Model(&driverSession).Exec(ctx)

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

func (d driverSessionRepository) DeleteByDriverId(ctx context.Context, driverId string) error {
	db := GetQueryContext(ctx)

	res, err := db.NewDelete().Model(&entity.DriverSession{}).Where("driver_id = ?", driverId).Exec(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return value.ErrNotFound
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

func NewDriverSessionRepository() driverSessionRepository {
	return driverSessionRepository{}
}
