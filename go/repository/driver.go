package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
)

// TODO(taekyeom) Driver repository에서 location 정보까지 같이 handling 해야 할까?
type DriverRepository interface {
	FindById(context.Context, string) (entity.Driver, error)
	FindByUserUniqueKey(context.Context, string) (entity.Driver, error)
	Create(context.Context, entity.Driver) error
	Update(context.Context, entity.Driver) error
	Delete(context.Context, entity.Driver) error
}

type driverRepository struct{}

func (u driverRepository) FindById(ctx context.Context, driverId string) (entity.Driver, error) {
	db := GetQueryContext(ctx)

	driver := entity.Driver{
		Id: driverId,
	}

	err := db.NewSelect().Model(&driver).WherePK().Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return entity.Driver{}, value.ErrDriverNotFound
	}
	if err != nil {
		return entity.Driver{}, fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return driver, nil
}

func (u driverRepository) FindByUserUniqueKey(ctx context.Context, userUniqueKey string) (entity.Driver, error) {
	db := GetQueryContext(ctx)

	driver := entity.Driver{}

	err := db.NewSelect().Model(&driver).Where("user_unique_key = ?", userUniqueKey).Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return entity.Driver{}, value.ErrDriverNotFound
	}
	if err != nil {
		return entity.Driver{}, fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return driver, nil
}

func (u driverRepository) Create(ctx context.Context, driver entity.Driver) error {
	db := GetQueryContext(ctx)

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

func (u driverRepository) Update(ctx context.Context, driver entity.Driver) error {
	db := GetQueryContext(ctx)

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

func (u driverRepository) Delete(ctx context.Context, driver entity.Driver) error {
	db := GetQueryContext(ctx)

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

func NewDriverRepository() driverRepository {
	return driverRepository{}
}
