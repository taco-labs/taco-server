package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/cridenour/go-postgis"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
)

type DriverLocationRepository interface {
	GetByDriverId(context.Context, string) (entity.DriverLocation, error)
	GetByRadius(context.Context, postgis.PointS, float64) ([]entity.DriverLocation, error)
	Upsert(context.Context, entity.DriverLocation) error
}

type driverLocationRepository struct{}

func (d driverLocationRepository) GetByDriverId(ctx context.Context, driverId string) (entity.DriverLocation, error) {
	db := GetQueryContext(ctx)

	driverLocation := entity.DriverLocation{}

	err := db.NewSelect().Model(&driverLocation).WherePK(driverId).Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return entity.DriverLocation{}, value.ErrDriverNotFound
	}
	if err != nil {
		return entity.DriverLocation{}, fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return driverLocation, nil
}

func (d driverLocationRepository) GetByRadius(ctx context.Context, point postgis.PointS, radiusMeter float64) ([]entity.DriverLocation, error) {
	// TODO (taekyeom) make query
	return []entity.DriverLocation{}, nil
}

func (d driverLocationRepository) Upsert(ctx context.Context, location entity.DriverLocation) error {
	db := GetQueryContext(ctx)

	_, err := db.NewInsert().
		Model(&location).
		On("CONFLICT (driver_id) UPDATE").
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
