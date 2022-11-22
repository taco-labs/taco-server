package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/uptrace/bun"
)

type driverLocationModel struct {
	bun.BaseModel `bun:"driver_location"`

	DriverId   string    `bun:"driver_id,pk"`
	EwkbHex    string    `bun:"location"`
	UpdateTime time.Time `bun:"update_time"`
}

type DriverLocationRepository interface {
	GetByDriverId(context.Context, bun.IDB, string) (entity.DriverLocation, error)
	Upsert(context.Context, bun.IDB, entity.DriverLocation) error
}

type driverLocationRepository struct{}

func (d driverLocationRepository) GetByDriverId(ctx context.Context, db bun.IDB, driverId string) (entity.DriverLocation, error) {
	driverLocation := driverLocationModel{
		DriverId: driverId,
	}

	err := db.NewSelect().Model(&driverLocation).WherePK().Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return entity.DriverLocation{}, value.ErrNotFound
	}
	if err != nil {
		return entity.DriverLocation{}, fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return DriverLocationFromModel(driverLocation)
}

func (d driverLocationRepository) Upsert(ctx context.Context, db bun.IDB, location entity.DriverLocation) error {
	model, err := DriverLocationToModel(location)
	if err != nil {
		return err
	}
	_, err = db.NewInsert().
		Model(&model).
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

func DriverLocationToModel(dto entity.DriverLocation) (driverLocationModel, error) {
	ewkbHex, err := dto.Location.ToEwkbHex()

	if err != nil {
		return driverLocationModel{}, err
	}

	return driverLocationModel{
		DriverId:   dto.DriverId,
		EwkbHex:    ewkbHex,
		UpdateTime: dto.UpdateTime,
	}, nil
}

func DriverLocationFromModel(model driverLocationModel) (entity.DriverLocation, error) {
	driverLocationEntity := entity.DriverLocation{
		DriverId:   model.DriverId,
		UpdateTime: model.UpdateTime,
	}

	if err := driverLocationEntity.Location.FromEwkbHex(model.EwkbHex); err != nil {
		return entity.DriverLocation{}, err
	}

	return driverLocationEntity, nil
}
