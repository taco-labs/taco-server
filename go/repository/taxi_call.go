package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/domain/value/enum"
	"github.com/taco-labs/taco/go/utils/slices"
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

	GetTicketById(context.Context, bun.IDB, string) (entity.TaxiCallTicket, error)
	GetLatestTicketByRequestId(context.Context, bun.IDB, string) (entity.TaxiCallTicket, error)
	CreateTicket(context.Context, bun.IDB, entity.TaxiCallTicket) error
	TicketExists(context.Context, bun.IDB, entity.TaxiCallTicket) (bool, error)
	DeleteTicketByRequestId(context.Context, bun.IDB, string) error
	GetDistributedCountByTicketId(context.Context, bun.IDB, string) (int, error)

	GetDriverTaxiCallContext(context.Context, bun.IDB, string) (entity.DriverTaxiCallContext, error)
	UpsertDriverTaxiCallContext(context.Context, bun.IDB, entity.DriverTaxiCallContext) error
	BulkUpsertDriverTaxiCallContext(context.Context, bun.IDB, []entity.DriverTaxiCallContext) error
	GetDriverTaxiCallContextByTicketId(context.Context, bun.IDB, string) ([]entity.DriverTaxiCallContext, error)

	GetDriverTaxiCallContextWithinRadius(context.Context, bun.IDB,
		value.Location, value.Location, int, string, time.Time) ([]entity.DriverTaxiCallContext, error)
}

type taxiCallRepository struct{}

func (t taxiCallRepository) BulkUpsertDriverTaxiCallContext(ctx context.Context, db bun.IDB,
	callContexts []entity.DriverTaxiCallContext) error {

	_, err := db.NewInsert().
		Model(&callContexts).
		On("CONFLICT (driver_id) DO UPDATE").
		Set("can_receive = EXCLUDED.can_receive").
		Set("last_received_request_ticket = EXCLUDED.last_received_request_ticket").
		Set("rejected_last_request_ticket = EXCLUDED.rejected_last_request_ticket").
		Set("last_receive_time = EXCLUDED.last_receive_time").
		Set("block_until = EXCLUDED.block_until").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return nil
}

func (t taxiCallRepository) GetDriverTaxiCallContextByTicketId(ctx context.Context, db bun.IDB, ticketId string) ([]entity.DriverTaxiCallContext, error) {
	resp := []entity.DriverTaxiCallContext{}

	err := db.NewSelect().Model(&resp).Where("last_received_request_ticket = ?", ticketId).Scan(ctx)

	if err != nil {
		return []entity.DriverTaxiCallContext{}, fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (t taxiCallRepository) GetTicketById(ctx context.Context, db bun.IDB, ticketId string) (entity.TaxiCallTicket, error) {
	resp := entity.TaxiCallTicket{}

	err := db.NewSelect().Model(&resp).
		Order("attempt DESC").
		Where("ticket_id = ?", ticketId).
		Limit(1).
		Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return entity.TaxiCallTicket{}, value.ErrNotFound
	}
	if err != nil {
		return entity.TaxiCallTicket{}, fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (t taxiCallRepository) GetLatestTicketByRequestId(ctx context.Context, db bun.IDB, taxiCallRequestId string) (entity.TaxiCallTicket, error) {
	resp := entity.TaxiCallTicket{}

	err := db.NewSelect().Model(&resp).
		Where("taxi_call_request_id = ?", taxiCallRequestId).
		Order("create_time DESC").
		Limit(1).
		Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return entity.TaxiCallTicket{}, value.ErrNotFound
	}
	if err != nil {
		return entity.TaxiCallTicket{}, fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (t taxiCallRepository) CreateTicket(ctx context.Context, db bun.IDB, ticket entity.TaxiCallTicket) error {
	res, err := db.NewInsert().Model(&ticket).Exec(ctx)

	if err != nil {
		return fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
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

func (t taxiCallRepository) TicketExists(ctx context.Context, db bun.IDB, ticket entity.TaxiCallTicket) (bool, error) {
	exists, err := db.NewSelect().Model(&ticket).WherePK().Exists(ctx)

	if err != nil {
		return false, fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	return exists, nil
}

func (t taxiCallRepository) DeleteTicketByRequestId(ctx context.Context, db bun.IDB, taxiCallRequestId string) error {
	_, err := db.NewDelete().Model((*entity.TaxiCallTicket)(nil)).Where("taxi_call_request_id = ?", taxiCallRequestId).Exec(ctx)

	if err != nil {
		return fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	return nil
}

func (t taxiCallRepository) GetDistributedCountByTicketId(ctx context.Context, db bun.IDB, ticketId string) (int, error) {
	var distributedCount int

	err := db.NewSelect().
		Model((*entity.TaxiCallTicket)(nil)).
		Where("ticket_id = ?", ticketId).
		ColumnExpr("SUM(distributed_count)").
		Scan(ctx, &distributedCount)

	if err != nil {
		return 0, fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	return distributedCount, nil
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

func (t taxiCallRepository) GetDriverTaxiCallContextWithinRadius(ctx context.Context, db bun.IDB,
	departure value.Location, arrival value.Location, raidus int, ticketId string, requestTime time.Time) ([]entity.DriverTaxiCallContext, error) {

	type tempModel struct {
		entity.DriverTaxiCallContext `bun:",extend"`

		Distance int    `bun:"distance"`
		EwkbHex  string `bun:"location"`
	}

	var resp []tempModel

	locationWithDistance := db.NewSelect().
		TableExpr("driver_location").
		ColumnExpr("driver_id").
		ColumnExpr("location").
		// TODO (taekyeom) Handle public schema search path...
		ColumnExpr("public.ST_DistanceSphere(location, public.ST_GeomFromText('POINT(? ?)',4326)) as distance",
			departure.Point.Longitude, departure.Point.Latitude)

	locationWithDistanceFiltered := db.NewSelect().
		TableExpr("driver_distance").
		ColumnExpr("driver_id").
		ColumnExpr("location").
		ColumnExpr("distance").
		Where("distance <= ?", raidus)

	driverServiceRegion := db.NewSelect().
		TableExpr("driver").
		Column("id").
		ColumnExpr("service_region").
		Where("service_region = ?", departure.Address.ServiceRegion).
		WhereOr("service_region = ?", arrival.Address.ServiceRegion)

	err := db.NewSelect().
		With("driver_distance", locationWithDistance).
		With("driver_distance_filtered", locationWithDistanceFiltered).
		With("driver_service_region", driverServiceRegion).
		Model(&resp).
		ColumnExpr("driver_taxi_call_context.*").
		ColumnExpr("location").
		ColumnExpr("CAST(distance AS int) as distance").
		Join("JOIN driver_distance_filtered AS t2 ON t2.driver_id = ?TableName.driver_id").
		Join("JOIN driver_service_region AS t3 ON t3.id = ?TableName.driver_id").
		Where("block_until is NULL or block_until < ?", requestTime).
		Where("can_receive").
		Where("last_received_request_ticket <> ? AND (rejected_last_request_ticket OR ? - last_receive_time > '10 seconds')", ticketId, requestTime).
		Order("distance").
		Limit(20). // TODO (taekyeom) To be smart limit..
		Scan(ctx)

	if err != nil {
		return []entity.DriverTaxiCallContext{}, fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	return slices.MapErr(resp, func(i tempModel) (entity.DriverTaxiCallContext, error) {
		if err := i.Location.FromEwkbHex(i.EwkbHex); err != nil {
			return entity.DriverTaxiCallContext{}, err
		}
		i.DriverTaxiCallContext.ToDepartureDistance = i.Distance

		return i.DriverTaxiCallContext, nil
	})
}

func (t taxiCallRepository) GetDriverTaxiCallContext(ctx context.Context, db bun.IDB, driverId string) (entity.DriverTaxiCallContext, error) {
	resp := struct {
		entity.DriverTaxiCallContext `bun:",extend"`

		EwkbHex string `bun:"location"`
	}{
		DriverTaxiCallContext: entity.DriverTaxiCallContext{
			DriverId: driverId,
		},
	}

	err := db.NewSelect().Model(&resp).
		ColumnExpr("driver_taxi_call_context.*").
		ColumnExpr("location").
		Join("JOIN driver_location").
		JoinOn("driver_taxi_call_context.driver_id = driver_location.driver_id").
		WherePK().Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return entity.DriverTaxiCallContext{}, value.ErrNotFound
	}
	if err != nil {
		return entity.DriverTaxiCallContext{}, fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	if err := resp.Location.FromEwkbHex(resp.EwkbHex); err != nil {
		return entity.DriverTaxiCallContext{}, err
	}

	return resp.DriverTaxiCallContext, nil
}

func (t taxiCallRepository) UpsertDriverTaxiCallContext(ctx context.Context, db bun.IDB, callContext entity.DriverTaxiCallContext) error {
	_, err := db.NewInsert().
		Model(&callContext).
		On("CONFLICT (driver_id) DO UPDATE").
		Set("can_receive = EXCLUDED.can_receive").
		Set("last_received_request_ticket = EXCLUDED.last_received_request_ticket").
		Set("rejected_last_request_ticket = EXCLUDED.rejected_last_request_ticket").
		Set("last_receive_time = EXCLUDED.last_receive_time").
		Set("block_until = EXCLUDED.block_until").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return nil
}

func NewTaxiCallRepository() *taxiCallRepository {
	return &taxiCallRepository{}
}
