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
}

type DriverLocationRepository interface {
	GetByDriverId(context.Context, bun.IDB, string) (entity.DriverLocation, error)
	GetByRadius(context.Context, bun.IDB, postgis.PointS, float64) ([]entity.DriverLocation, error)
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

func (d driverLocationRepository) GetByRadius(ctx context.Context, db bun.IDB, point postgis.PointS, radiusMeter float64) ([]entity.DriverLocation, error) {
	// TODO (taekyeom) make query
	return []entity.DriverLocation{}, nil
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
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return nil
}

func NewDriverLocationRepository() driverLocationRepository {
	return driverLocationRepository{}
}

func DriverLocationToModel(dto entity.DriverLocation) (driverLocationModel, error) {
	geomPoint := geom.NewPoint(geom.XY).
		MustSetCoords([]float64{dto.Location.Longitude, dto.Location.Latitude}).
		SetSRID(value.SRID_SPHERE)

	ewkbHex, err := ewkbhex.Encode(geomPoint, ewkbhex.NDR)
	if err != nil {
		return driverLocationModel{}, fmt.Errorf("%w: error while encode location: %v", value.ErrInternal, err)
	}

	return driverLocationModel{
		DriverId: dto.DriverId,
		Location: ewkbHex,
	}, nil
}

func DriverLocationFromModel(model driverLocationModel) (entity.DriverLocation, error) {
	point, err := ewkbhex.Decode(model.Location)
	if err != nil {
		return entity.DriverLocation{}, fmt.Errorf("%w: error while decode location: %v", value.ErrInternal, err)
	}
	if point.Layout() != geom.XY {
		return entity.DriverLocation{}, fmt.Errorf("%w: invalid location data", value.ErrDBInternal)
	}
	coords := point.FlatCoords()
	return entity.DriverLocation{
		DriverId: model.DriverId,
		Location: value.Point{
			Longitude: coords[0],
			Latitude:  coords[1],
		},
	}, nil
}
