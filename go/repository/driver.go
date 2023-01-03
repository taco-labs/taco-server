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

	ListNotActivatedDriver(context.Context, bun.IDB, string, int) ([]entity.DriverDto, string, error)

	CreateDriverRegistrationNumber(context.Context, bun.IDB, entity.DriverResidentRegistrationNumber) error

	GetDriverCarProfile(context.Context, bun.IDB, string) (entity.DriverCarProfile, error)
	ListDriverCarProfileByDriverId(context.Context, bun.IDB, string) ([]entity.DriverCarProfile, error)
	CreateDriverCarProfile(context.Context, bun.IDB, entity.DriverCarProfile) error
	UpdateDriverCarProfile(context.Context, bun.IDB, entity.DriverCarProfile) error
	DeleteDriverCarProfile(context.Context, bun.IDB, entity.DriverCarProfile) error
}

type driverRepository struct{}

func (u driverRepository) FindById(ctx context.Context, db bun.IDB, driverId string) (entity.DriverDto, error) {
	driver := entity.DriverDto{
		Id: driverId,
	}

	err := db.NewSelect().Model(&driver).WherePK().Relation("CarProfile").Scan(ctx)

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

	err := db.NewSelect().Model(&driver).Relation("CarProfile").Where("user_unique_key = ?", userUniqueKey).Scan(ctx)

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

func (u driverRepository) ListNotActivatedDriver(ctx context.Context, db bun.IDB, pageToken string, count int) ([]entity.DriverDto, string, error) {
	var resp []entity.DriverDto

	selectExpr := db.NewSelect().Model(&resp).Where("not active").Relation("CarProfile").Order("create_time DESC").Limit(count)

	if pageToken != "" {
		subQ := db.NewSelect().Model((*entity.DriverDto)(nil)).Column("create_time").Where("id = ?", pageToken)
		selectExpr = selectExpr.Where("create_time < (?)", subQ)
	}

	err := selectExpr.Scan(ctx)

	if err != nil {
		return []entity.DriverDto{}, "", fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	resultCount := len(resp)
	if resultCount == 0 {
		return resp, pageToken, nil
	}
	return resp, resp[resultCount-1].Id, nil
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

func (u driverRepository) GetDriverCarProfile(ctx context.Context, db bun.IDB, id string) (entity.DriverCarProfile, error) {
	resp := entity.DriverCarProfile{Id: id}

	err := db.NewSelect().Model(&resp).ColumnExpr("*").ColumnExpr("EXISTS (select 1 from driver where driver.car_profile_id = ?TableName.id) as selected").WherePK().Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.DriverCarProfile{}, value.ErrNotFound
	}
	if err != nil {
		return entity.DriverCarProfile{}, fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (u driverRepository) ListDriverCarProfileByDriverId(ctx context.Context, db bun.IDB, driverId string) ([]entity.DriverCarProfile, error) {
	resp := []entity.DriverCarProfile{}

	err := db.NewSelect().Model(&resp).ColumnExpr("*").ColumnExpr("EXISTS (select 1 from driver where driver.car_profile_id = ?TableName.id) as selected").Where("driver_id = ?", driverId).Scan(ctx)

	if err != nil {
		return []entity.DriverCarProfile{}, fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (u driverRepository) CreateDriverCarProfile(ctx context.Context, db bun.IDB, carProfile entity.DriverCarProfile) error {
	res, err := db.NewInsert().Model(&carProfile).Exec(ctx)

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

func (u driverRepository) UpdateDriverCarProfile(ctx context.Context, db bun.IDB, carProfile entity.DriverCarProfile) error {
	res, err := db.NewUpdate().Model(&carProfile).WherePK().Exec(ctx)

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

func (u driverRepository) DeleteDriverCarProfile(ctx context.Context, db bun.IDB, carProfile entity.DriverCarProfile) error {
	res, err := db.NewDelete().Model(&carProfile).WherePK().Exec(ctx)

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
