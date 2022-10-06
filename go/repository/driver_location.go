package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/cridenour/go-postgis"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/uptrace/bun"
)

type DriverLocationRepository interface {
	GetByDriverId(context.Context, bun.IDB, string) (entity.DriverLocation, error)
	GetByRadius(context.Context, bun.IDB, postgis.PointS, float64) ([]entity.DriverLocation, error)
	Upsert(context.Context, bun.IDB, entity.DriverLocation) error
}

type driverLocationRepository struct{}

func (d driverLocationRepository) GetByDriverId(ctx context.Context, db bun.IDB, driverId string) (entity.DriverLocation, error) {
	driverLocation := entity.DriverLocation{}

	err := db.NewSelect().Model(&driverLocation).WherePK(driverId).Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return entity.DriverLocation{}, value.ErrNotFound
	}
	if err != nil {
		return entity.DriverLocation{}, fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return driverLocation, nil
}

func (d driverLocationRepository) GetByRadius(ctx context.Context, db bun.IDB, point postgis.PointS, radiusMeter float64) ([]entity.DriverLocation, error) {
	// TODO (taekyeom) make query
	return []entity.DriverLocation{}, nil
}

func (d driverLocationRepository) Upsert(ctx context.Context, db bun.IDB, location entity.DriverLocation) error {
	_, err := db.NewInsert().
		Model(&location).
		On("CONFLICT (driver_id) DO UPDATE").
		Set("location = EXCLUDED.location").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return nil
}

func NewDriverLocationRepository() driverLocationRepository {
	return driverLocationRepository{}
}
