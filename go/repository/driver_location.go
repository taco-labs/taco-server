package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/cridenour/go-postgis"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkbhex"
	"github.com/uptrace/bun"
)

type driverLocationModel struct {
	bun.BaseModel `bun:"driver_location"`

	Location string `bun:"location"`
	DriverId string `bun:"driver_id,pk"`
	OnDuty   bool   `bun:"on_duty"`
}

type DriverLocationRepository interface {
	GetByDriverId(context.Context, bun.IDB, string) (entity.DriverLocationDto, error)
	GetByRadius(context.Context, bun.IDB, postgis.PointS, float64) ([]entity.DriverLocationDto, error)
	Upsert(context.Context, bun.IDB, entity.DriverLocationDto) error
}

type driverLocationRepository struct{}

func (d driverLocationRepository) GetByDriverId(ctx context.Context, db bun.IDB, driverId string) (entity.DriverLocationDto, error) {
	driverLocation := driverLocationModel{
		DriverId: driverId,
	}

	err := db.NewSelect().Model(&driverLocation).WherePK().Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return entity.DriverLocationDto{}, value.ErrNotFound
	}
	if err != nil {
		return entity.DriverLocationDto{}, fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return DriverLocationFromModel(driverLocation)
}

func (d driverLocationRepository) GetByRadius(ctx context.Context, db bun.IDB, point postgis.PointS, radiusMeter float64) ([]entity.DriverLocationDto, error) {
	// TODO (taekyeom) make query
	return []entity.DriverLocationDto{}, nil
}

func (d driverLocationRepository) Upsert(ctx context.Context, db bun.IDB, location entity.DriverLocationDto) error {
	model, err := DriverLocationToModel(location)
	if err != nil {
		return err
	}
	_, err = db.NewInsert().
		Model(&model).
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

func DriverLocationToModel(dto entity.DriverLocationDto) (driverLocationModel, error) {
	ewkbHex, err := ewkbhex.Encode(dto.Location, ewkbhex.NDR)
	if err != nil {
		return driverLocationModel{}, fmt.Errorf("%w: error while encode location: %v", value.ErrInternal, err)
	}

	return driverLocationModel{
		DriverId: dto.DriverId,
		Location: ewkbHex,
		OnDuty:   dto.OnDuty,
	}, nil
}

func DriverLocationFromModel(model driverLocationModel) (entity.DriverLocationDto, error) {
	point, err := ewkbhex.Decode(model.Location)
	if err != nil {
		return entity.DriverLocationDto{}, fmt.Errorf("%w: error while decode location: %v", value.ErrInternal, err)
	}
	if point.Layout() != geom.XY {
		return entity.DriverLocationDto{}, fmt.Errorf("%w: invalid location data", value.ErrDBInternal)
	}
	coords := point.FlatCoords()
	return entity.NewDriverLocation(model.DriverId, coords[1], coords[0], model.OnDuty), nil
}
