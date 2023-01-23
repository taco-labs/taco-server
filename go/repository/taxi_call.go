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
	"github.com/taco-labs/taco/go/utils"
	"github.com/uptrace/bun"
)

type TaxiCallRepository interface {
	// TaxiCallRequest
	GetById(context.Context, bun.IDB, string) (entity.TaxiCallRequest, error)
	GetLatestByUserId(context.Context, bun.IDB, string) (entity.TaxiCallRequest, error)
	GetLatestByDriverId(context.Context, bun.IDB, string) (entity.TaxiCallRequest, error)
	ListByUserId(context.Context, bun.IDB, string, string, int) ([]entity.TaxiCallRequest, string, error)
	ListByDriverId(context.Context, bun.IDB, string, string, int) ([]entity.TaxiCallRequest, string, error)
	Create(context.Context, bun.IDB, entity.TaxiCallRequest) error
	Update(context.Context, bun.IDB, entity.TaxiCallRequest) error

	// Route
	CreateToDepartureRoute(context.Context, bun.IDB, entity.TaxiCallToDepartureRoute) error
	CreateToArrivalRoute(context.Context, bun.IDB, entity.TaxiCallToArrivalRoute) error

	GetActiveRequestIds(context.Context, bun.IDB) ([]string, error)

	// Ticket
	GetTicketById(context.Context, bun.IDB, string) (entity.TaxiCallTicket, error)
	GetLatestTicketByRequestId(context.Context, bun.IDB, string) (entity.TaxiCallTicket, error)
	CreateTicket(context.Context, bun.IDB, entity.TaxiCallTicket) error
	TicketExists(context.Context, bun.IDB, entity.TaxiCallTicket) (bool, error)
	DeleteTicketByRequestId(context.Context, bun.IDB, string) error
	GetDistributedCountByTicketId(context.Context, bun.IDB, string) (int, error)

	// DriverContext
	GetDriverTaxiCallContext(context.Context, bun.IDB, string) (entity.DriverTaxiCallContext, error)
	UpsertDriverTaxiCallContext(context.Context, bun.IDB, entity.DriverTaxiCallContext) error
	BulkUpsertDriverTaxiCallContext(context.Context, bun.IDB, []entity.DriverTaxiCallContext) error
	ActivateTicketNonAcceptedDriverContext(context.Context, bun.IDB, string, string) error

	GetDriverTaxiCallContextWithinRadius(context.Context, bun.IDB,
		value.Location, value.Location, int, []int, string, time.Time) ([]entity.DriverTaxiCallContext, error)

	ListDriverTaxiCallContextInRadius(context.Context, bun.IDB, value.Point, string, int) ([]entity.DriverTaxiCallContextWithInfo, error)

	// Tag
	ListDriverDenyTaxiCallTag(context.Context, bun.IDB, string) ([]entity.DriverDenyTaxiCallTag, error)
	CreateDriverDenyTaxiCallTag(context.Context, bun.IDB, entity.DriverDenyTaxiCallTag) error
	DeleteDriverDenyTaxiCallTag(context.Context, bun.IDB, entity.DriverDenyTaxiCallTag) error
}

type taxiCallRepository struct{}

func (t taxiCallRepository) GetById(ctx context.Context, db bun.IDB, taxiCallRequestId string) (entity.TaxiCallRequest, error) {
	resp := entity.TaxiCallRequest{
		Id: taxiCallRequestId,
	}

	err := db.NewSelect().
		Model(&resp).
		WherePK().
		Relation("ToDepartureRoute").
		Relation("ToArrivalRoute").
		Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return resp, value.ErrNotFound
	}
	if err != nil {
		return resp, fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (t taxiCallRepository) GetLatestByUserId(ctx context.Context, db bun.IDB, userId string) (entity.TaxiCallRequest, error) {
	resp := entity.TaxiCallRequest{}

	err := db.NewSelect().
		Model(&resp).
		Where("user_id = ?", userId).
		Relation("ToDepartureRoute").
		Relation("ToArrivalRoute").
		OrderExpr("create_time DESC").Limit(1).Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return resp, value.ErrNotFound
	}
	if err != nil {
		return resp, fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (t taxiCallRepository) GetLatestByDriverId(ctx context.Context, db bun.IDB, driverId string) (entity.TaxiCallRequest, error) {
	resp := entity.TaxiCallRequest{}

	err := db.NewSelect().
		Model(&resp).
		Where("driver_id = ?", driverId).
		Where("taxi_call_state <> ?", enum.TaxiCallState_MOCK_CALL_ACCEPTED).
		Relation("ToDepartureRoute").
		Relation("ToArrivalRoute").
		OrderExpr("create_time DESC").Limit(1).Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return resp, value.ErrNotFound
	}
	if err != nil {
		return resp, fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (t taxiCallRepository) ListByUserId(ctx context.Context, db bun.IDB, userId string, pageToken string, count int) ([]entity.TaxiCallRequest, string, error) {
	var resp []entity.TaxiCallRequest

	selectExpr := db.NewSelect().
		Model(&resp).
		Where("user_id = ?", userId).
		Relation("ToDepartureRoute").
		Relation("ToArrivalRoute").
		Order("create_time DESC").
		Limit(count)

	if pageToken != "" {
		subQ := db.NewSelect().Model((*entity.TaxiCallRequest)(nil)).Column("create_time").Where("id = ?", pageToken)
		selectExpr = selectExpr.Where("create_time < (?)", subQ)
	}

	err := selectExpr.Scan(ctx)
	if err != nil {
		return resp, "", fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	resultCount := len(resp)
	if resultCount == 0 {
		return resp, pageToken, nil
	}
	return resp, resp[resultCount-1].Id, nil
}

func (t taxiCallRepository) ListByDriverId(ctx context.Context, db bun.IDB, driverId string, pageToken string, count int) ([]entity.TaxiCallRequest, string, error) {
	var resp []entity.TaxiCallRequest

	selectExpr := db.NewSelect().
		Model(&resp).
		Where("driver_id = ?", driverId).
		Relation("ToDepartureRoute").
		Relation("ToArrivalRoute").
		Order("create_time DESC").
		Where("taxi_call_state <> ?", enum.TaxiCallState_MOCK_CALL_ACCEPTED).
		Limit(count)

	if pageToken != "" {
		subQ := db.NewSelect().Model((*entity.TaxiCallRequest)(nil)).Column("create_time").Where("id = ?", pageToken)
		selectExpr = selectExpr.Where("create_time < (?)", subQ)
	}

	err := selectExpr.Scan(ctx)
	if err != nil {
		return resp, "", fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	resultCount := len(resp)
	if resultCount == 0 {
		return resp, pageToken, nil
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

func (t taxiCallRepository) CreateToDepartureRoute(ctx context.Context, db bun.IDB, toDepartureRoute entity.TaxiCallToDepartureRoute) error {
	res, err := db.NewInsert().Model(&toDepartureRoute).Exec(ctx)

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

func (t taxiCallRepository) CreateToArrivalRoute(ctx context.Context, db bun.IDB, toArrivalRoute entity.TaxiCallToArrivalRoute) error {
	res, err := db.NewInsert().Model(&toArrivalRoute).Exec(ctx)

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

func (t taxiCallRepository) BulkUpsertDriverTaxiCallContext(ctx context.Context, db bun.IDB,
	callContexts []entity.DriverTaxiCallContext) error {

	_, err := db.NewInsert().
		Model(&callContexts).
		On("CONFLICT (driver_id) DO UPDATE").
		Set("can_receive = EXCLUDED.can_receive").
		Set("last_received_request_ticket = EXCLUDED.last_received_request_ticket").
		Set("rejected_last_request_ticket = EXCLUDED.rejected_last_request_ticket").
		Set("last_receive_time = EXCLUDED.last_receive_time").
		Set("to_departure_distance = EXCLUDED.to_departure_distance").
		Set("block_until = EXCLUDED.block_until").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return nil
}

func (t taxiCallRepository) ActivateTicketNonAcceptedDriverContext(ctx context.Context, db bun.IDB, receivedDriverId string, ticketId string) error {
	_, err := db.NewUpdate().Model((*entity.DriverTaxiCallContext)(nil)).
		Where("driver_id <> ?", receivedDriverId).
		Where("last_received_request_ticket = ?", ticketId).
		Set("rejected_last_request_ticket = ?", true).Exec(ctx)

	if err != nil {
		return fmt.Errorf("%w: error from db %v", value.ErrDBInternal, err)
	}

	return nil
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

func (t taxiCallRepository) GetDriverTaxiCallContextWithinRadius(ctx context.Context, db bun.IDB,
	departure value.Location, arrival value.Location, radius int, requestTagIds []int,
	ticketId string, requestTime time.Time) ([]entity.DriverTaxiCallContext, error) {

	var resp []entity.DriverTaxiCallContext

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
		Where("distance <= ?", radius)

	driverServiceRegion := db.NewSelect().
		TableExpr("driver").
		Column("id").
		ColumnExpr("service_region").
		Where("service_region = ?", departure.Address.ServiceRegion).
		WhereOr("service_region = ?", arrival.Address.ServiceRegion)

	driverTaxiCallContexts := db.NewSelect().
		With("driver_distance", locationWithDistance).
		With("driver_distance_filtered", locationWithDistanceFiltered).
		With("driver_service_region", driverServiceRegion).
		Model(&resp).
		ColumnExpr("driver_taxi_call_context.*").
		ColumnExpr("location").
		ColumnExpr("CAST(distance AS int) as to_departure_distance").
		Join("JOIN driver_distance_filtered AS t2 ON t2.driver_id = ?TableName.driver_id").
		Join("JOIN driver_service_region AS t3 ON t3.id = ?TableName.driver_id").
		Where("block_until is NULL or block_until < ?", requestTime).
		Where("can_receive").
		Where("last_received_request_ticket <> ? AND (rejected_last_request_ticket OR ? - last_receive_time > '10 seconds')", ticketId, requestTime).
		Order("distance").
		Limit(20) // TODO (taekyeom) To be smart limit..

	if len(requestTagIds) > 0 {
		driverTaxiCallContexts = driverTaxiCallContexts.
			Where("NOT EXISTS (SELECT 1 FROM driver_deny_taxi_call_tag WHERE driver_id = ?TableName.driver_id AND tag_id IN (?))", bun.In(requestTagIds))
	}

	err := driverTaxiCallContexts.Scan(ctx)

	if err != nil {
		return []entity.DriverTaxiCallContext{}, fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (t taxiCallRepository) ListDriverTaxiCallContextInRadius(ctx context.Context, db bun.IDB, point value.Point, serviceRegion string, radius int) ([]entity.DriverTaxiCallContextWithInfo, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	var resp []entity.DriverTaxiCallContextWithInfo

	locationWithDistance := db.NewSelect().
		TableExpr("driver_location").
		ColumnExpr("driver_id").
		ColumnExpr("location").
		// TODO (taekyeom) Handle public schema search path...
		ColumnExpr("public.ST_DistanceSphere(location, public.ST_GeomFromText('POINT(? ?)',4326)) as distance",
			point.Longitude, point.Latitude)

	locationWithDistanceFiltered := db.NewSelect().
		TableExpr("driver_distance").
		ColumnExpr("driver_id").
		ColumnExpr("location").
		ColumnExpr("distance").
		Where("distance <= ?", radius)

	driverServiceRegion := db.NewSelect().
		TableExpr("driver").
		Column("id").
		ColumnExpr("first_name").
		ColumnExpr("last_name").
		ColumnExpr("app_version").
		ColumnExpr("service_region").
		Where("service_region = ?", serviceRegion).
		WhereOr("service_region = ?", serviceRegion)

	driverTaxiCallContexts := db.NewSelect().
		With("driver_distance", locationWithDistance).
		With("driver_distance_filtered", locationWithDistanceFiltered).
		With("driver_service_region", driverServiceRegion).
		Model(&resp).
		ColumnExpr("driver_taxi_call_context.*").
		ColumnExpr("location").
		ColumnExpr("CAST(distance AS int) as to_departure_distance").
		ColumnExpr("first_name").
		ColumnExpr("last_name").
		ColumnExpr("app_version").
		Join("JOIN driver_distance_filtered AS t2 ON t2.driver_id = ?TableName.driver_id").
		Join("JOIN driver_service_region AS t3 ON t3.id = ?TableName.driver_id").
		Where("block_until is NULL or block_until < ?", requestTime).
		Where("can_receive").
		Order("distance").
		Limit(20) // TODO (taekyeom) To be smart limit..

	err := driverTaxiCallContexts.Scan(ctx)

	if err != nil {
		return []entity.DriverTaxiCallContextWithInfo{}, fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (t taxiCallRepository) GetDriverTaxiCallContext(ctx context.Context, db bun.IDB, driverId string) (entity.DriverTaxiCallContext, error) {
	resp := entity.DriverTaxiCallContext{
		DriverId: driverId,
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

	return resp, nil
}

func (t taxiCallRepository) UpsertDriverTaxiCallContext(ctx context.Context, db bun.IDB, callContext entity.DriverTaxiCallContext) error {
	_, err := db.NewInsert().
		Model(&callContext).
		On("CONFLICT (driver_id) DO UPDATE").
		Set("can_receive = EXCLUDED.can_receive").
		Set("last_received_request_ticket = EXCLUDED.last_received_request_ticket").
		Set("rejected_last_request_ticket = EXCLUDED.rejected_last_request_ticket").
		Set("last_receive_time = EXCLUDED.last_receive_time").
		Set("to_departure_distance = EXCLUDED.to_departure_distance").
		Set("block_until = EXCLUDED.block_until").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return nil
}

func (t taxiCallRepository) ListDriverDenyTaxiCallTag(ctx context.Context, db bun.IDB, driverId string) ([]entity.DriverDenyTaxiCallTag, error) {
	resp := []entity.DriverDenyTaxiCallTag{}

	err := db.NewSelect().Model(&resp).Where("driver_id = ?", driverId).Order("tag_id").Scan(ctx)

	if err != nil {
		return []entity.DriverDenyTaxiCallTag{}, fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (t taxiCallRepository) CreateDriverDenyTaxiCallTag(ctx context.Context, db bun.IDB, denyTag entity.DriverDenyTaxiCallTag) error {
	res, err := db.NewInsert().Model(&denyTag).Exec(ctx)

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

func (t taxiCallRepository) DeleteDriverDenyTaxiCallTag(ctx context.Context, db bun.IDB, denyTag entity.DriverDenyTaxiCallTag) error {
	res, err := db.NewDelete().Model(&denyTag).WherePK().Exec(ctx)

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

func NewTaxiCallRepository() *taxiCallRepository {
	return &taxiCallRepository{}
}
