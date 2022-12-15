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

// TODO(taekyeom) Driver repository에서 location 정보까지 같이 handling 해야 할까?
type DriverRepository interface {
	FindById(context.Context, bun.IDB, string) (entity.DriverDto, error)
	FindByUserUniqueKey(context.Context, bun.IDB, string) (entity.DriverDto, error)
	Create(context.Context, bun.IDB, entity.DriverDto) error
	Update(context.Context, bun.IDB, entity.DriverDto) error
	Delete(context.Context, bun.IDB, entity.DriverDto) error

	CreateDriverRegistrationNumber(context.Context, bun.IDB, entity.DriverResidentRegistrationNumber) error
}

type driverRepository struct{}

func (u driverRepository) FindById(ctx context.Context, db bun.IDB, driverId string) (entity.DriverDto, error) {
	driver := entity.DriverDto{
		Id: driverId,
	}

	err := db.NewSelect().Model(&driver).WherePK().Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return entity.DriverDto{}, value.ErrDriverNotFound
	}
	if err != nil {
		return entity.DriverDto{}, fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return driver, nil
}

func (u driverRepository) FindByUserUniqueKey(ctx context.Context, db bun.IDB, userUniqueKey string) (entity.DriverDto, error) {
	driver := entity.DriverDto{}

	err := db.NewSelect().Model(&driver).Where("user_unique_key = ?", userUniqueKey).Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return entity.DriverDto{}, value.ErrDriverNotFound
	}
	if err != nil {
		return entity.DriverDto{}, fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return driver, nil
}

func (u driverRepository) Create(ctx context.Context, db bun.IDB, driver entity.DriverDto) error {
	res, err := db.NewInsert().Model(&driver).Exec(ctx)

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

func (u driverRepository) Update(ctx context.Context, db bun.IDB, driver entity.DriverDto) error {
	res, err := db.NewUpdate().Model(&driver).WherePK().Exec(ctx)

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

func (u driverRepository) Delete(ctx context.Context, db bun.IDB, driver entity.DriverDto) error {
	res, err := db.NewDelete().Model(&driver).WherePK().Exec(ctx)

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

func (u driverRepository) CreateDriverRegistrationNumber(ctx context.Context, db bun.IDB, registrationNumber entity.DriverResidentRegistrationNumber) error {
	res, err := db.NewInsert().Model(&registrationNumber).Exec(ctx)

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

func NewDriverRepository() *driverRepository {
	return &driverRepository{}
}
