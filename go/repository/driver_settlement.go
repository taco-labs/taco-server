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

	GetDriverTotalSettlement(context.Context, bun.IDB, string) (entity.DriverTotalSettlement, error)
	UpdateTotalDriverSettlement(context.Context, bun.IDB, string, int) error

	ListDriverSettlementHistory(context.Context, bun.IDB, string, time.Time, int) ([]entity.DriverSettlementHistory, time.Time, error)
	CreateDriverSettlementHistory(context.Context, bun.IDB, entity.DriverSettlementHistory) error

	AggregateDriverRequestableSettlement(context.Context, bun.IDB, string, time.Time) (int, error)
	BatchDeleteDriverSettlementRequest(context.Context, bun.IDB, string, time.Time) error

	GetInflightSettlementTransferById(context.Context, bun.IDB, string) (entity.DriverInflightSettlementTransfer, error)
	GetInflightSettlementTransferByDriverId(context.Context, bun.IDB, string) (entity.DriverInflightSettlementTransfer, error)
	CreateInflightSettlementTransfer(context.Context, bun.IDB, entity.DriverInflightSettlementTransfer) error
	UpdateInflightSettlementTransfer(context.Context, bun.IDB, entity.DriverInflightSettlementTransfer) error
	DeleteInflightSettlementTransfer(context.Context, bun.IDB, entity.DriverInflightSettlementTransfer) error

	GetFailedSettlementTransferById(context.Context, bun.IDB, string) (entity.DriverFailedSettlementTransfer, error)
	CreateFailedSettlementTransfer(context.Context, bun.IDB, entity.DriverFailedSettlementTransfer) error
	DeleteFailedSettlementTransfer(context.Context, bun.IDB, entity.DriverFailedSettlementTransfer) error

	GetDriverPromotionSettlementReward(context.Context, bun.IDB, string) (entity.DriverPromotionSettlementReward, error)
	CreateDriverPromotionSettlementReward(context.Context, bun.IDB, entity.DriverPromotionSettlementReward) error
	UpdateDriverPromotionSettlementReward(context.Context, bun.IDB, entity.DriverPromotionSettlementReward) error
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

func (d driverSettlementRepository) GetDriverTotalSettlement(ctx context.Context, db bun.IDB, driverId string) (entity.DriverTotalSettlement, error) {
	resp := entity.DriverTotalSettlement{
		DriverId: driverId,
	}

	err := db.NewSelect().Model(&resp).WherePK().Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.DriverTotalSettlement{}, value.ErrNotFound
	}
	if err != nil {
		return entity.DriverTotalSettlement{}, fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (d driverSettlementRepository) UpdateTotalDriverSettlement(ctx context.Context, db bun.IDB, driverId string, amount int) error {
	upsertStatement := db.NewInsert().
		Model(&entity.DriverTotalSettlement{DriverId: driverId, TotalAmount: amount}).
		On("CONFLICT (driver_id) DO UPDATE")

	if amount > 0 {
		upsertStatement = upsertStatement.Set("total_amount = driver_total_settlement.total_amount + EXCLUDED.total_amount")
	} else {
		upsertStatement = upsertStatement.Set("total_amount = driver_total_settlement.total_amount - EXCLUDED.total_amount")
	}

	_, err := upsertStatement.Exec(ctx)

	if err != nil {
		return fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return nil
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
		return []entity.DriverSettlementHistory{}, pageToken, nil
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

func (d driverSettlementRepository) AggregateDriverRequestableSettlement(ctx context.Context, db bun.IDB, driverId string, requestTime time.Time) (int, error) {
	var amount int

	err := db.NewSelect().Model(&entity.DriverSettlementRequest{}).
		Where("create_time < ?", requestTime).
		Where("driver_id = ?", driverId).
		ColumnExpr("sum(amount)").
		Scan(ctx, &amount)

	if err != nil {
		return 0, fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return amount, nil
}

func (d driverSettlementRepository) BatchDeleteDriverSettlementRequest(ctx context.Context, db bun.IDB, driverId string, requestTime time.Time) error {
	_, err := db.NewDelete().Model(&entity.DriverSettlementRequest{}).
		Where("create_time < ?", requestTime).
		Where("driver_id = ?", driverId).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return nil
}

func NewDriverSettlementRepository() *driverSettlementRepository {
	return &driverSettlementRepository{}
}

func (d driverSettlementRepository) GetInflightSettlementTransferById(ctx context.Context, db bun.IDB, transferId string) (entity.DriverInflightSettlementTransfer, error) {
	resp := entity.DriverInflightSettlementTransfer{
		TransferId: transferId,
	}

	err := db.NewSelect().Model(&resp).WherePK().Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.DriverInflightSettlementTransfer{}, value.ErrNotFound
	}
	if err != nil {
		return entity.DriverInflightSettlementTransfer{}, fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (d driverSettlementRepository) GetInflightSettlementTransferByDriverId(ctx context.Context, db bun.IDB, driverId string) (entity.DriverInflightSettlementTransfer, error) {
	resp := entity.DriverInflightSettlementTransfer{}

	err := db.NewSelect().Model(&resp).Where("driver_id = ?", driverId).Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.DriverInflightSettlementTransfer{}, value.ErrNotFound
	}
	if err != nil {
		return entity.DriverInflightSettlementTransfer{}, fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (d driverSettlementRepository) CreateInflightSettlementTransfer(ctx context.Context, db bun.IDB, inflightSettlementTransfer entity.DriverInflightSettlementTransfer) error {
	res, err := db.NewInsert().Model(&inflightSettlementTransfer).Exec(ctx)

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

func (d driverSettlementRepository) UpdateInflightSettlementTransfer(ctx context.Context, db bun.IDB, inflightSettlementTransfer entity.DriverInflightSettlementTransfer) error {
	res, err := db.NewUpdate().Model(&inflightSettlementTransfer).WherePK().Exec(ctx)

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

func (d driverSettlementRepository) DeleteInflightSettlementTransfer(ctx context.Context, db bun.IDB, inflightSettlementTransfer entity.DriverInflightSettlementTransfer) error {
	res, err := db.NewDelete().Model(&inflightSettlementTransfer).WherePK().Exec(ctx)

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

func (d driverSettlementRepository) GetFailedSettlementTransferById(ctx context.Context, db bun.IDB, transferId string) (entity.DriverFailedSettlementTransfer, error) {
	resp := entity.DriverFailedSettlementTransfer{
		TransferId: transferId,
	}

	err := db.NewSelect().Model(&resp).WherePK().Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.DriverFailedSettlementTransfer{}, value.ErrNotFound
	}
	if err != nil {
		return entity.DriverFailedSettlementTransfer{}, fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (d driverSettlementRepository) CreateFailedSettlementTransfer(ctx context.Context, db bun.IDB, failedSettlementTransfer entity.DriverFailedSettlementTransfer) error {
	res, err := db.NewInsert().Model(&failedSettlementTransfer).Exec(ctx)

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

func (d driverSettlementRepository) DeleteFailedSettlementTransfer(ctx context.Context, db bun.IDB, failedSettlementTransfer entity.DriverFailedSettlementTransfer) error {
	res, err := db.NewDelete().Model(&failedSettlementTransfer).WherePK().Exec(ctx)

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

func (d driverSettlementRepository) GetDriverPromotionSettlementReward(ctx context.Context, db bun.IDB, driverId string) (entity.DriverPromotionSettlementReward, error) {
	resp := entity.DriverPromotionSettlementReward{
		DriverId: driverId,
	}

	err := db.NewSelect().Model(&resp).WherePK().Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.DriverPromotionSettlementReward{}, value.ErrNotFound
	}
	if err != nil {
		return entity.DriverPromotionSettlementReward{}, fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (d driverSettlementRepository) CreateDriverPromotionSettlementReward(ctx context.Context, db bun.IDB, promotionReward entity.DriverPromotionSettlementReward) error {
	res, err := db.NewInsert().Model(&promotionReward).Exec(ctx)

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

func (d driverSettlementRepository) UpdateDriverPromotionSettlementReward(ctx context.Context, db bun.IDB, promotionReward entity.DriverPromotionSettlementReward) error {
	res, err := db.NewUpdate().Model(&promotionReward).WherePK().Exec(ctx)

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
