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

type DriverLocationRepository interface {
	GetByDriverId(context.Context, bun.IDB, string) (entity.DriverLocation, error)
	Upsert(context.Context, bun.IDB, entity.DriverLocation) error
}

type driverLocationRepository struct{}

func (d driverLocationRepository) GetByDriverId(ctx context.Context, db bun.IDB, driverId string) (entity.DriverLocation, error) {
	driverLocation := entity.DriverLocation{
		DriverId: driverId,
	}

	err := db.NewSelect().Model(&driverLocation).WherePK().Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return entity.DriverLocation{}, value.ErrNotFound
	}
	if err != nil {
		return entity.DriverLocation{}, fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return driverLocation, nil
}

func (d driverLocationRepository) Upsert(ctx context.Context, db bun.IDB, location entity.DriverLocation) error {
	_, err := db.NewInsert().
		Model(&location).
		On("CONFLICT (driver_id) DO UPDATE").
		Set("location = EXCLUDED.location").
		Set("update_time = EXCLUDED.update_time").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return nil
}

func NewDriverLocationRepository() *driverLocationRepository {
	return &driverLocationRepository{}
}
