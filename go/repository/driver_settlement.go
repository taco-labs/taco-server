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

type DriverSettlementRepository interface {
	GetDriverSettlementRequest(context.Context, bun.IDB, string) (entity.DriverSettlementRequest, error)
	CreateDriverSettlementRequest(context.Context, bun.IDB, entity.DriverSettlementRequest) error

	GetDriverExpectedSettlement(context.Context, bun.IDB, string) (entity.DriverExpectedSettlement, error)
	UpdateExpectedDriverSettlement(context.Context, bun.IDB, string, int) error

	GetDriverSettlementHistory(context.Context, bun.IDB, string, time.Time, time.Time) (entity.DriverSettlementHistory, error)
	ListDriverSettlementHistory(context.Context, bun.IDB, string, time.Time, int) ([]entity.DriverSettlementHistory, time.Time, error)
	CreateDriverSettlementHistory(context.Context, bun.IDB, entity.DriverSettlementHistory) error
}

type driverSettlementRepository struct{}

func (d driverSettlementRepository) GetDriverSettlementRequest(ctx context.Context, db bun.IDB, taxiCallRquestId string) (entity.DriverSettlementRequest, error) {
	resp := entity.DriverSettlementRequest{
		TaxiCallRequestId: taxiCallRquestId,
	}

	err := db.NewSelect().Model(&resp).WherePK().Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.DriverSettlementRequest{}, value.ErrNotFound
	}
	if err != nil {
		return entity.DriverSettlementRequest{}, fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (d driverSettlementRepository) CreateDriverSettlementRequest(ctx context.Context, db bun.IDB, settlementRequest entity.DriverSettlementRequest) error {
	res, err := db.NewInsert().Model(&settlementRequest).Exec(ctx)

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

func (d driverSettlementRepository) GetDriverExpectedSettlement(ctx context.Context, db bun.IDB, driverId string) (entity.DriverExpectedSettlement, error) {
	resp := entity.DriverExpectedSettlement{
		DriverId: driverId,
	}

	err := db.NewSelect().Model(&resp).WherePK().Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.DriverExpectedSettlement{}, value.ErrNotFound
	}
	if err != nil {
		return entity.DriverExpectedSettlement{}, fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (d driverSettlementRepository) UpdateExpectedDriverSettlement(ctx context.Context, db bun.IDB, driverId string, amount int) error {
	upsertStatement := db.NewInsert().
		Model(&entity.DriverExpectedSettlement{DriverId: driverId, ExpectedAmount: amount}).
		On("CONFLICT (driver_id) DO UPDATE")

	if amount > 0 {
		upsertStatement = upsertStatement.Set("expected_amount = driver_expected_settlement.expected_amount + EXCLUDED.expected_amount")
	} else {
		upsertStatement = upsertStatement.Set("expected_amount = driver_expected_settlement.expected_amount - EXCLUDED.expected_amount")
	}

	_, err := upsertStatement.Exec(ctx)

	if err != nil {
		return fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return nil
}

func (d driverSettlementRepository) GetDriverSettlementHistory(ctx context.Context, db bun.IDB, driverId string, periodStart, periodEnd time.Time) (entity.DriverSettlementHistory, error) {
	resp := entity.DriverSettlementHistory{
		DriverId:              driverId,
		SettlementPeriodStart: periodStart,
		SettlementPeriodEnd:   periodEnd,
	}

	err := db.NewSelect().Model(&resp).WherePK().Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.DriverSettlementHistory{}, value.ErrNotFound
	}
	if err != nil {
		return entity.DriverSettlementHistory{}, fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (d driverSettlementRepository) ListDriverSettlementHistory(ctx context.Context, db bun.IDB, driverId string, pageToken time.Time, count int) ([]entity.DriverSettlementHistory, time.Time, error) {
	resp := []entity.DriverSettlementHistory{}

	query := db.NewSelect().Model(&resp).
		Order("create_time DESC").
		Limit(count)

	if !pageToken.IsZero() {
		query = query.Where("create_time < ?", pageToken)
	}

	err := query.Scan(ctx)

	if err != nil {
		return []entity.DriverSettlementHistory{}, time.Time{}, fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	if len(resp) == 0 {
		return []entity.DriverSettlementHistory{}, time.Time{}, nil
	}

	return resp, resp[len(resp)-1].CreateTime, nil
}

func (d driverSettlementRepository) CreateDriverSettlementHistory(ctx context.Context, db bun.IDB, settlementHistory entity.DriverSettlementHistory) error {
	res, err := db.NewInsert().Model(&settlementHistory).Exec(ctx)

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

func NewDriverSettlementRepository() *driverSettlementRepository {
	return &driverSettlementRepository{}
}
