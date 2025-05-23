package taxicall

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/event/command"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/domain/value/enum"
	"github.com/taco-labs/taco/go/utils"
	"github.com/uptrace/bun"
)

func (t taxicallApp) ActivateDriverContext(ctx context.Context, driverId string) error {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	return t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		taxiCallContext, err := t.repository.taxiCallRequest.GetDriverTaxiCallContext(ctx, i, driverId)
		if err != nil && !errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("app.taxiCall.ActivateDriverContext: error while get last call request: %w", err)
		}
		if errors.Is(err, value.ErrNotFound) {
			taxiCallContext = entity.NewEmptyDriverTaxiCallContext(driverId, true, requestTime)
		}
		taxiCallContext.CanReceive = true

		if err := t.repository.taxiCallRequest.UpsertDriverTaxiCallContext(ctx, i, taxiCallContext); err != nil {
			return fmt.Errorf("app.taxiCall.ActivateDriverContext: error while upsert driver taxi call context: %w", err)
		}

		return nil
	})
}

func (t taxicallApp) DeactivateDriverContext(ctx context.Context, driverId string) error {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	return t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		lastTaxiCallRequest, err := t.repository.taxiCallRequest.GetLatestByDriverId(ctx, i, driverId)
		if err != nil && !errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("app.taxiCall.DeactivateDriverContext: error while get last call request: %w", err)
		}
		if lastTaxiCallRequest.CurrentState.Active() {
			return fmt.Errorf("app.taxiCall.DeactivateDriverContext: active taxi call request exists: %w", value.ErrActiveTaxiCallRequestExists)
		}

		taxiCallContext, err := t.repository.taxiCallRequest.GetDriverTaxiCallContext(ctx, i, driverId)
		if err != nil && !errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("app.taxiCall.DeactivateDriverContext: error while get last call request: %w", err)
		}
		if errors.Is(err, value.ErrNotFound) {
			taxiCallContext = entity.NewEmptyDriverTaxiCallContext(driverId, false, requestTime)
		}
		taxiCallContext.CanReceive = false

		if err := t.repository.taxiCallRequest.UpsertDriverTaxiCallContext(ctx, i, taxiCallContext); err != nil {
			return fmt.Errorf("app.taxiCall.DeactivateDriverContext: error while upsert driver taxi call context: %w", err)
		}

		return nil
	})
}

func (t taxicallApp) UpdateDriverLocation(ctx context.Context, req request.DriverLocationUpdateRequest) error {
	return t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		driverLocationDto := entity.DriverLocation{
			DriverId: req.DriverId,
			Location: value.Point{
				Latitude:  req.Latitude,
				Longitude: req.Longitude,
			},
		}

		if err := t.repository.driverLocation.Upsert(ctx, i, driverLocationDto); err != nil {
			return fmt.Errorf("app.taxiCall.UpdateDriverLocation: error while update driver location:\n%w", err)
		}

		return nil
	})
}

func (t taxicallApp) ListDriverTaxiCallRequest(ctx context.Context, req request.ListDriverTaxiCallRequest) ([]entity.TaxiCallRequest, string, error) {
	var taxiCallRequests []entity.TaxiCallRequest
	var err error
	var pageToken string

	err = t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		taxiCallRequests, pageToken, err = t.repository.taxiCallRequest.ListByDriverId(ctx, i, req.DriverId, req.PageToken, req.Count)
		if err != nil {
			return fmt.Errorf("app.taxiCall.ListDriverTaxiCallRequest: error while get taxi call requests:\n%w", err)
		}
		return nil
	})

	if err != nil {
		return []entity.TaxiCallRequest{}, "", nil
	}

	return taxiCallRequests, pageToken, nil
}

func (t taxicallApp) LatestDriverTaxiCallRequest(ctx context.Context, driverId string) (entity.TaxiCallRequest, error) {
	var latestTaxiCallRequest entity.TaxiCallRequest
	var err error

	err = t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		latestTaxiCallRequest, err = t.repository.taxiCallRequest.GetLatestByDriverId(ctx, i, driverId)
		if err != nil {
			return fmt.Errorf("app.driver.LatestDriverTaxiCallRequest: error while get latest taxi call:\n%w", err)
		}
		return nil
	})

	if err != nil {
		return entity.TaxiCallRequest{}, err
	}

	return latestTaxiCallRequest, nil
}

// TODO (taekyeom) Remove it later!!
func (t taxicallApp) ForceAcceptTaxiCallRequest(ctx context.Context, driverId, callRequestId string) error {
	return t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		ticket, err := t.repository.taxiCallRequest.GetLatestTicketByRequestId(ctx, i, callRequestId)
		if err != nil {
			return fmt.Errorf("app.taxiCall.ForceAcceptTaxiCallRequest: error while find latest ticket by request id: %w", err)
		}

		// TODO(taeykeom) Do we need check on duty & last call request?
		driverTaxiCallContext, err := t.repository.taxiCallRequest.GetDriverTaxiCallContext(ctx, i, driverId)
		if err != nil {
			return fmt.Errorf("app.taxiCall.ForceAcceptTaxiCallRequest: error while get taxi call context:%w", err)
		}

		driverTaxiCallContext.CanReceive = false
		driverTaxiCallContext.LastReceivedRequestTicket = ticket.Id
		if err := t.repository.taxiCallRequest.UpsertDriverTaxiCallContext(ctx, i, driverTaxiCallContext); err != nil {
			return fmt.Errorf("app.taxiCall.ForceAcceptTaxiCallRequest: error while upsert taxi call context: %w", value.ErrInvalidOperation)
		}

		return t.AcceptTaxiCallRequest(ctx, driverId, ticket.Id)
	})
}

// TODO (taekyeom) Add route between driver location & departure?
func (t taxicallApp) AcceptTaxiCallRequest(ctx context.Context, driverId string, ticketId string) error {
	requestTime := utils.GetRequestTimeOrNow(ctx)
	var taxiCallRequest entity.TaxiCallRequest
	var driverTaxiCallContext entity.DriverTaxiCallContext
	var err error

	err = t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		// TODO(taeykeom) Do we need check on duty & last call request?
		driverTaxiCallContext, err = t.repository.taxiCallRequest.GetDriverTaxiCallContext(ctx, i, driverId)
		if err != nil {
			return fmt.Errorf("app.taxxiCall.AcceptTaxiCallRequest: error while get taxi call context:%w", err)
		}
		if driverTaxiCallContext.LastReceivedRequestTicket != ticketId {
			return fmt.Errorf("app.taxxiCall.AcceptTaxiCallRequest: invalid ticket id: %w", value.ErrInvalidOperation)
		}

		driverTaxiCallContext.CanReceive = false
		if err := t.repository.taxiCallRequest.UpsertDriverTaxiCallContext(ctx, i, driverTaxiCallContext); err != nil {
			return fmt.Errorf("app.taxxiCall.AcceptTaxiCallRequest: error while upsert taxi call context: %w", value.ErrInvalidOperation)
		}

		// TODO(taekyeom) ticket과 현재 ticket이 다른 경우, 돈을 더 받는것도 괜찮을까?
		ticket, err := t.repository.taxiCallRequest.GetLatestTicketByRequestId(ctx, i, driverTaxiCallContext.LastReceivedRequestTicket)
		if err != nil {
			return fmt.Errorf("app.taxxiCall.AcceptTaxiCallRequest: error while get taxi call ticket:%w", err)
		}

		taxiCallRequest, err = t.repository.taxiCallRequest.GetById(ctx, i, driverTaxiCallContext.LastReceivedRequestTicket)
		if err != nil {
			return fmt.Errorf("app.taxxiCall.AcceptTaxiCallRequest: error while get taxi call request:%w", err)
		}
		if !taxiCallRequest.CurrentState.Requested() {
			return fmt.Errorf("app.taxxiCall.AcceptTaxiCallRequest: already expired taxi call request:%w", value.ErrAlreadyExpiredCallRequest)
		}

		if err := taxiCallRequest.UpdateState(requestTime, enum.TaxiCallState_DRIVER_TO_DEPARTURE); err != nil {
			return fmt.Errorf("app.taxxiCall.AcceptTaxiCallRequest: invalid state change:%w", err)
		}

		taxiCallRequest.AdditionalPrice = ticket.AdditionalPrice
		taxiCallRequest.DriverId = sql.NullString{
			Valid:  true,
			String: driverId,
		}
		if err := t.repository.taxiCallRequest.Update(ctx, i, taxiCallRequest); err != nil {
			return fmt.Errorf("app.taxxiCall.AcceptTaxiCallRequest: error while update taxi call request :%w", err)
		}

		userCmd := command.NewUserTaxiCallNotificationCommand(taxiCallRequest, entity.TaxiCallTicket{}, entity.DriverTaxiCallContext{})
		if err := t.repository.event.BatchCreate(ctx, i, []entity.Event{userCmd}); err != nil {
			return fmt.Errorf("app.taxxiCall.AcceptTaxiCallRequest: error while create event: %w", err)
		}

		processMessage := command.TaxiCallProcessMessage{
			TaxiCallRequestId:   taxiCallRequest.Id,
			TaxiCallState:       string(taxiCallRequest.CurrentState),
			EventTime:           taxiCallRequest.UpdateTime,
			DesiredScheduleTime: taxiCallRequest.UpdateTime,
		}

		if err := t.repository.event.BatchCreate(ctx, i, []entity.Event{processMessage.ToEvent()}); err != nil {
			return fmt.Errorf("app.taxiCall.AcceptTaxiCallRequest: error while create taxi call process event: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (t taxicallApp) RejectTaxiCallRequest(ctx context.Context, driverId string, ticketId string) error {
	return t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		// TODO(taeykeom) Do we need check on duty & last call request?
		driverTaxiCallContext, err := t.repository.taxiCallRequest.GetDriverTaxiCallContext(ctx, i, driverId)
		if err != nil {
			return fmt.Errorf("app.taxxiCall.RejectTaxiCallRequest: error while get taxi call context:%w", err)
		}
		if driverTaxiCallContext.LastReceivedRequestTicket != ticketId {
			return fmt.Errorf("app.taxxiCall.RejectTaxiCallRequest: invalid ticket id: %w", value.ErrInvalidOperation)
		}

		driverTaxiCallContext.RejectedLastRequestTicket = true
		if err := t.repository.taxiCallRequest.UpsertDriverTaxiCallContext(ctx, i, driverTaxiCallContext); err != nil {
			return fmt.Errorf("app.taxxiCall.RejectTaxiCallRequest: error while upsert taxi call context: %w", value.ErrInvalidOperation)
		}

		return nil
	})
}

func (d taxicallApp) DriverToArrival(ctx context.Context, driverId string, callRequestId string) error {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	return d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		taxiCallRequest, err := d.repository.taxiCallRequest.GetById(ctx, i, callRequestId)
		if err != nil {
			return fmt.Errorf("app.taxxiCall.DriverToArrival: error while get taxi call request: %w", err)
		}

		if taxiCallRequest.DriverId.String != driverId {
			return fmt.Errorf("app.taxiCall.DriverToArrival: unauthorized access: %w", value.ErrUnAuthorized)
		}

		if err := taxiCallRequest.UpdateState(requestTime, enum.TaxiCallState_DRIVER_TO_ARRIVAL); err != nil {
			return fmt.Errorf("app.taxxiCall.DriverToArrival: invalid state change: %w", err)
		}

		if err := d.repository.taxiCallRequest.Update(ctx, i, taxiCallRequest); err != nil {
			return fmt.Errorf("app.taxxiCall.DriverToArrival: error while update taxi call request: %w", err)
		}

		// TODO (taekyeom) send push?

		return nil
	})
}

func (t taxicallApp) DoneTaxiCallRequest(ctx context.Context, driverId string, req request.DoneTaxiCallRequest) error {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	var taxiCallRequest entity.TaxiCallRequest
	var err error

	err = t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		taxiCallRequest, err = t.repository.taxiCallRequest.GetById(ctx, i, req.TaxiCallRequestId)
		if err != nil {
			return fmt.Errorf("app.taxxiCall.DoneTaxiCallRequest: error while get taxi call request: %w", err)
		}

		if taxiCallRequest.CurrentState.Complete() {
			// TODO (taeykeom) change error code
			return fmt.Errorf("app.taxxiCall.DoneTaxiCallRequest: already completed call request: %w", value.ErrAlreadyExpiredCallRequest)
		}

		if taxiCallRequest.DriverId.String != driverId {
			return fmt.Errorf("app.taxxiCall.DoneTaxiCallRequest: forbidden access: %w", value.ErrUnAuthorized)
		}

		if err := taxiCallRequest.UpdateState(requestTime, enum.TaxiCallState_DONE); err != nil {
			return fmt.Errorf("app.taxxiCall.DoneTaxiCallRequest: invalid state change:%w", err)
		}

		taxiCallRequest.BasePrice = req.BasePrice
		taxiCallRequest.UpdateTime = requestTime

		if err := t.repository.taxiCallRequest.Update(ctx, i, taxiCallRequest); err != nil {
			return fmt.Errorf("app.taxxiCall.DoneTaxiCallRequest: error while update taxi call request :%w", err)
		}

		driverTaxiCallContext, err := t.repository.taxiCallRequest.GetDriverTaxiCallContext(ctx, i, driverId)
		if err != nil {
			return fmt.Errorf("app.taxxiCall.RejectTaxiCallRequest: error while get taxi call context:%w", err)
		}

		driverTaxiCallContext.CanReceive = true
		if err := t.repository.taxiCallRequest.UpsertDriverTaxiCallContext(ctx, i, driverTaxiCallContext); err != nil {
			return fmt.Errorf("app.taxxiCall.DoneTaxiCallRequest: error while upsert taxi call context: %w", err)
		}

		taxiCallSettlement := entity.DriverTaxiCallSettlement{
			TaxiCallRequestId: taxiCallRequest.Id,
			SettlementDone:    false,
		}
		if err := t.repository.taxiCallRequest.CreateDriverTaxiCallSettlement(ctx, i, taxiCallSettlement); err != nil {
			return fmt.Errorf("app.taxxiCall.DoneTaxiCallRequest: error while create taxi call settlement: %w", err)
		}

		processMessage := command.TaxiCallProcessMessage{
			TaxiCallRequestId:   taxiCallRequest.Id,
			TaxiCallState:       string(taxiCallRequest.CurrentState),
			EventTime:           taxiCallRequest.UpdateTime,
			DesiredScheduleTime: taxiCallRequest.UpdateTime,
		}

		if err := t.repository.event.BatchCreate(ctx, i, []entity.Event{processMessage.ToEvent()}); err != nil {
			return fmt.Errorf("app.taxiCall.DoneTaxiCallRequest: error while create taxi call process event: %w", err)
		}

		// TODO (taekyeom) Do payment command

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
