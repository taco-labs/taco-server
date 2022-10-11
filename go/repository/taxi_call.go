package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/domain/value/enum"
	"github.com/uptrace/bun"
)

type TaxiCallRepository interface {
	GetById(context.Context, bun.IDB, string) (entity.TaxiCallRequest, error)
	GetLatestByUserId(context.Context, bun.IDB, string) (entity.TaxiCallRequest, error)
	GetLatestByDriverId(context.Context, bun.IDB, string) (entity.TaxiCallRequest, error)
	ListByUserId(context.Context, bun.IDB, string, string, int) ([]entity.TaxiCallRequest, string, error)
	ListByDriverId(context.Context, bun.IDB, string, string, int) ([]entity.TaxiCallRequest, string, error)
	Create(context.Context, bun.IDB, entity.TaxiCallRequest) error
	Update(context.Context, bun.IDB, entity.TaxiCallRequest) error

	GetActiveRequestIds(context.Context, bun.IDB) ([]string, error)

	GetLatestTicketByRequestId(context.Context, bun.IDB, string) (entity.TaxiCallTicket, error)
	UpsertTicket(context.Context, bun.IDB, entity.TaxiCallTicket) error
	DeleteTicketByRequestId(context.Context, bun.IDB, string) error

	// GetLastReceivedTicket(context.Context, bun.IDB, string) (entity.TaxiCallLastReceivedTicket, error)
	// UpsertLastReceivedTicket(context.Context, bun.IDB, entity.TaxiCallLastReceivedTicket) error
}

type taxiCallRepository struct{}

func (t taxiCallRepository) GetLatestTicketByRequestId(ctx context.Context, db bun.IDB, requestId string) (entity.TaxiCallTicket, error) {
	resp := entity.TaxiCallTicket{}

	err := db.NewSelect().Model(&resp).
		Where("taxi_call_request_id = ?", requestId).
		Order("create_time DESC").
		Limit(1).
		Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return entity.TaxiCallTicket{}, value.ErrNotFound
	}
	if err != nil {
		return resp, fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (t taxiCallRepository) UpsertTicket(ctx context.Context, db bun.IDB, ticket entity.TaxiCallTicket) error {
	_, err := db.NewInsert().
		Model(&ticket).
		On("CONFLICT (id) DO UPDATE").
		Set("taxi_call_request_id = EXCLUDED.taxi_call_request_id").
		Set("attempt = EXCLUDED.attempt").
		Set("additional_price = EXCLUDED.additional_price").
		Set("create_time = EXCLUDED.create_time").
		Set("update_time = EXCLUDED.update_time").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return nil
}

func (t taxiCallRepository) DeleteTicketByRequestId(ctx context.Context, db bun.IDB, requestId string) error {
	_, err := db.NewDelete().Model((*entity.TaxiCallTicket)(nil)).Where("taxi_call_request_id = ?", requestId).Exec(ctx)

	if err != nil {
		return fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	return nil
}

func (t taxiCallRepository) GetById(ctx context.Context, db bun.IDB, taxiCallRequestId string) (entity.TaxiCallRequest, error) {
	resp := entity.TaxiCallRequest{
		Id: taxiCallRequestId,
	}

	err := db.NewSelect().Model(&resp).WherePK().Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return entity.TaxiCallRequest{}, value.ErrNotFound
	}
	if err != nil {
		return resp, fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (t taxiCallRepository) GetLatestByUserId(ctx context.Context, db bun.IDB, userId string) (entity.TaxiCallRequest, error) {
	resp := entity.TaxiCallRequest{}

	err := db.NewSelect().Model(&resp).Where("user_id = ?", userId).OrderExpr("create_time DESC").Limit(1).Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return entity.TaxiCallRequest{}, value.ErrNotFound
	}
	if err != nil {
		return resp, fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (t taxiCallRepository) GetLatestByDriverId(ctx context.Context, db bun.IDB, driverId string) (entity.TaxiCallRequest, error) {
	resp := entity.TaxiCallRequest{}

	err := db.NewSelect().Model(&resp).Where("driver_id = ?", driverId).OrderExpr("create_time DESC").Limit(1).Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return entity.TaxiCallRequest{}, value.ErrNotFound
	}
	if err != nil {
		return resp, fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (t taxiCallRepository) ListByUserId(ctx context.Context, db bun.IDB, userId string, pageToken string, count int) ([]entity.TaxiCallRequest, string, error) {
	resp := []entity.TaxiCallRequest{}

	selectExpr := db.NewSelect().
		Model(&resp).
		Where("user_id = ?", userId).
		Order("create_time DESC").
		Limit(count)

	if pageToken != "" {
		subQ := db.NewSelect().Model((*entity.TaxiCallRequest)(nil)).Column("create_time").Where("id = ?", pageToken)
		selectExpr.Where("create_time < (?)", subQ)
	}

	err := selectExpr.Scan(ctx)
	if err != nil {
		return resp, "", fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	resultCount := len(resp)
	if resultCount == 0 {
		return resp, "", nil
	}
	return resp, resp[resultCount-1].Id, nil
}

func (t taxiCallRepository) GetActiveRequestIds(ctx context.Context, db bun.IDB) ([]string, error) {
	resp := []string{}

	err := db.NewSelect().Model((*entity.TaxiCallRequest)(nil)).
		Column("id").
		Where("taxi_call_state IN (?)", bun.In([]enum.TaxiCallState{
			enum.TaxiCallState_Requested,
			enum.TaxiCallState_DRIVER_TO_DEPARTURE,
			enum.TaxiCallState_DRIVER_TO_ARRIVAL,
		})).Scan(ctx, &resp)
	if err != nil {
		return []string{}, fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (t taxiCallRepository) ListByDriverId(ctx context.Context, db bun.IDB, driverId string, pageToken string, count int) ([]entity.TaxiCallRequest, string, error) {
	resp := []entity.TaxiCallRequest{}

	selectExpr := db.NewSelect().
		Model(&resp).
		Where("driver_id = ?", driverId).
		Order("create_time DESC").
		Limit(count)

	if pageToken != "" {
		subQ := db.NewSelect().Model((*entity.TaxiCallRequest)(nil)).Column("create_time").Where("id = ?", pageToken)
		selectExpr.Where("create_time < (?)", subQ)
	}

	err := selectExpr.Scan(ctx)
	if err != nil {
		return resp, "", fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	resultCount := len(resp)
	if resultCount == 0 {
		return resp, "", nil
	}
	return resp, resp[resultCount-1].Id, nil
}

func (t taxiCallRepository) Create(ctx context.Context, db bun.IDB, taxiCallRequest entity.TaxiCallRequest) error {
	_, err := db.NewInsert().Model(&taxiCallRequest).Exec(ctx)

	// TODO (taekyeom) handle already exists
	if err != nil {
		return fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	return nil
}

func (t taxiCallRepository) Update(ctx context.Context, db bun.IDB, taxiCallRequest entity.TaxiCallRequest) error {
	res, err := db.NewUpdate().Model(&taxiCallRequest).WherePK().Exec(ctx)
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

func NewTaxiCallRepository() taxiCallRepository {
	return taxiCallRepository{}
}
