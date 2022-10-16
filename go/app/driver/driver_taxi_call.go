package driver

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/event/command"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/domain/value/enum"
	"github.com/taco-labs/taco/go/utils"
	"github.com/uptrace/bun"
)

func (d driverApp) ListTaxiCallRequest(ctx context.Context, req request.ListDriverTaxiCallRequest) ([]entity.TaxiCallRequest, string, error) {
	var taxiCallRequests []entity.TaxiCallRequest
	var pageToken string
	var err error

	err = d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		// TODO (settlement?)
		taxiCallRequests, pageToken, err = d.repository.taxiCallRequest.ListByDriverId(ctx, i, req.DriverId, req.PageToken, req.Count)
		if err != nil {
			return fmt.Errorf("app.driver.ListTaxiCallRequest: error while get taxi call requests:%w", err)
		}

		return nil
	})

	if err != nil {
		return []entity.TaxiCallRequest{}, "", err
	}

	return taxiCallRequests, pageToken, nil
}

func (d driverApp) GetLatestTaxiCallRequest(ctx context.Context, driverId string) (entity.TaxiCallRequest, error) {
	var latestTaxiCallRequest entity.TaxiCallRequest
	var err error

	err = d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		latestTaxiCallRequest, err = d.repository.taxiCallRequest.GetLatestByDriverId(ctx, i, driverId)
		if err != nil {
			return fmt.Errorf("app.driver.GetLatestTaxiCallRequest: error while get latest taxi call:\n%w", err)
		}
		return nil
	})

	if err != nil {
		return entity.TaxiCallRequest{}, err
	}

	return latestTaxiCallRequest, nil
}

// TODO (taekyeom) Remove it later!!
func (d driverApp) ForceAcceptTaxiCallRequest(ctx context.Context, driverId, callRequestId string) error {
	requestTime := utils.GetRequestTimeOrNow(ctx)
	var taxiCallRequest entity.TaxiCallRequest
	var driverTaxiCallContext entity.DriverTaxiCallContext
	var err error

	err = d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		ticket, err := d.repository.taxiCallRequest.GetLatestTicketByRequestId(ctx, i, callRequestId)
		if err != nil {
			return fmt.Errorf("app.Driver.ForceAcceptTaxiCallRequest: error while find latest ticket by request id: %w", err)
		}

		// TODO(taeykeom) Do we need check on duty & last call request?
		driverTaxiCallContext, err = d.repository.taxiCallRequest.GetDriverTaxiCallContext(ctx, i, driverId)
		if err != nil {
			return fmt.Errorf("app.Driver.ForceAcceptTaxiCallRequest: error while get taxi call context:%w", err)
		}

		driverTaxiCallContext.CanReceive = false
		driverTaxiCallContext.LastReceivedRequestTicket = ticket.Id
		if err := d.repository.taxiCallRequest.UpsertDriverTaxiCallContext(ctx, i, driverTaxiCallContext); err != nil {
			return fmt.Errorf("app.Driver.ForceAcceptTaxiCallRequest: error while upsert taxi call context: %w", value.ErrInvalidOperation)
		}

		taxiCallRequest, err = d.repository.taxiCallRequest.GetById(ctx, i, callRequestId)
		if err != nil {
			return fmt.Errorf("app.Driver.ForceAcceptTaxiCallRequest: error while get taxi requestt:%w", err)
		}
		if !taxiCallRequest.CurrentState.Requested() {
			return fmt.Errorf("app.Driver.ForceAcceptTaxiCallRequest: already expired taxi call request:%w", value.ErrAlreadyExpiredCallRequest)
		}

		if err := taxiCallRequest.UpdateState(requestTime, enum.TaxiCallState_DRIVER_TO_DEPARTURE); err != nil {
			return fmt.Errorf("app.Driver.ForceAcceptTaxiCallRequest: invalid state change:%w", err)
		}

		taxiCallRequest.AdditionalPrice = ticket.AdditionalPrice
		taxiCallRequest.DriverId = sql.NullString{
			Valid:  true,
			String: driverId,
		}
		if err := d.repository.taxiCallRequest.Update(ctx, i, taxiCallRequest); err != nil {
			return fmt.Errorf("app.Driver.ForceAcceptTaxiCallRequest: error while update taxi call request :%w", err)
		}

		userCmd := command.NewUserTaxiCallNotificationCommand(taxiCallRequest, ticket, driverTaxiCallContext)
		if err := d.repository.event.BatchCreate(ctx, i, []entity.Event{userCmd}); err != nil {
			return fmt.Errorf("app.Driver.ForceAcceptTaxiCallRequest: error while create event: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (d driverApp) DriverToArrival(ctx context.Context, callRequestId string) error {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	return d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		taxiCallRequest, err := d.repository.taxiCallRequest.GetById(ctx, i, callRequestId)
		if err != nil {
			return fmt.Errorf("app.Driver.DriverToArrival: error while get taxi call request: %w", err)
		}
		if err := taxiCallRequest.UpdateState(requestTime, enum.TaxiCallState_DRIVER_TO_ARRIVAL); err != nil {
			return fmt.Errorf("app.Driver.DriverToArrival: invalid state change: %w", err)
		}

		if err := d.repository.taxiCallRequest.Update(ctx, i, taxiCallRequest); err != nil {
			return fmt.Errorf("app.Driver.DriverToArrival: error while update taxi call request: %w", err)
		}

		// TODO (taekyeom) send push?

		return nil
	})
}

func (d driverApp) AcceptTaxiCallRequest(ctx context.Context, ticketId string) error {
	driverId := utils.GetDriverId(ctx)
	requestTime := utils.GetRequestTimeOrNow(ctx)
	var taxiCallRequest entity.TaxiCallRequest
	var driverTaxiCallContext entity.DriverTaxiCallContext
	var err error

	// TODO (taekyeom) Add route between driver location & departure

	err = d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		// TODO(taeykeom) Do we need check on duty & last call request?
		driverTaxiCallContext, err = d.repository.taxiCallRequest.GetDriverTaxiCallContext(ctx, i, driverId)
		if err != nil {
			return fmt.Errorf("app.Driver.AcceptTaxiCallRequest: error while get taxi call context:%w", err)
		}
		if driverTaxiCallContext.LastReceivedRequestTicket != ticketId {
			return fmt.Errorf("app.Driver.AcceptTaxiCallRequest: invalid ticket id: %w", value.ErrInvalidOperation)
		}

		driverTaxiCallContext.CanReceive = false
		if err := d.repository.taxiCallRequest.UpsertDriverTaxiCallContext(ctx, i, driverTaxiCallContext); err != nil {
			return fmt.Errorf("app.Driver.AcceptTaxiCallRequest: error while upsert taxi call context: %w", value.ErrInvalidOperation)
		}

		// TODO(taekyeom) ticket과 현재 ticket이 다른 경우, 돈을 더 받는것도 괜찮을까?
		ticket, err := d.repository.taxiCallRequest.GetLatestTicketByRequestId(ctx, i, driverTaxiCallContext.LastReceivedRequestTicket)
		if err != nil {
			return fmt.Errorf("app.Driver.AcceptTaxiCallRequest: error while get taxi call ticket:%w", err)
		}

		taxiCallRequest, err = d.repository.taxiCallRequest.GetById(ctx, i, driverTaxiCallContext.LastReceivedRequestTicket)
		if err != nil {
			return fmt.Errorf("app.Driver.AcceptTaxiCallRequest: error while get taxi call request:%w", err)
		}
		if !taxiCallRequest.CurrentState.Requested() {
			return fmt.Errorf("app.Driver.AcceptTaxiCallRequest: already expired taxi call request:%w", value.ErrAlreadyExpiredCallRequest)
		}

		if err := taxiCallRequest.UpdateState(requestTime, enum.TaxiCallState_DRIVER_TO_DEPARTURE); err != nil {
			return fmt.Errorf("app.Driver.AcceptTaxiCallRequest: invalid state change:%w", err)
		}

		taxiCallRequest.AdditionalPrice = ticket.AdditionalPrice
		taxiCallRequest.DriverId = sql.NullString{
			Valid:  true,
			String: driverId,
		}
		if err := d.repository.taxiCallRequest.Update(ctx, i, taxiCallRequest); err != nil {
			return fmt.Errorf("app.Driver.AcceptTaxiCallRequest: error while update taxi call request :%w", err)
		}

		userCmd := command.NewUserTaxiCallNotificationCommand(taxiCallRequest, entity.TaxiCallTicket{}, entity.DriverTaxiCallContext{})
		if err := d.repository.event.BatchCreate(ctx, i, []entity.Event{userCmd}); err != nil {
			return fmt.Errorf("app.Driver.AcceptTaxiCallRequest: error while create event: %w", err)
		}

		return nil
	})

	// TODO (taekyeom) send push message to user
	if err != nil {
		return err
	}

	return nil
}

func (d driverApp) RejectTaxiCallRequest(ctx context.Context, ticketId string) error {
	driverId := utils.GetDriverId(ctx)

	return d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		// TODO(taeykeom) Do we need check on duty & last call request?
		driverTaxiCallContext, err := d.repository.taxiCallRequest.GetDriverTaxiCallContext(ctx, i, driverId)
		if err != nil {
			return fmt.Errorf("app.Driver.RejectTaxiCallRequest: error while get taxi call context:%w", err)
		}
		if driverTaxiCallContext.LastReceivedRequestTicket != ticketId {
			return fmt.Errorf("app.Driver.RejectTaxiCallRequest: invalid ticket id: %w", value.ErrInvalidOperation)
		}

		driverTaxiCallContext.RejectedLastRequestTicket = true
		if err := d.repository.taxiCallRequest.UpsertDriverTaxiCallContext(ctx, i, driverTaxiCallContext); err != nil {
			return fmt.Errorf("app.Driver.RejectTaxiCallRequest: error while upsert taxi call context: %w", value.ErrInvalidOperation)
		}

		return nil
	})
}

func (d driverApp) DoneTaxiCallRequest(ctx context.Context, req request.DoneTaxiCallRequest) error {
	driverId := utils.GetDriverId(ctx)
	requestTime := utils.GetRequestTimeOrNow(ctx)

	var taxiCallRequest entity.TaxiCallRequest
	var err error

	err = d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		taxiCallRequest, err = d.repository.taxiCallRequest.GetById(ctx, i, req.TaxiCallRequestId)
		if err != nil {
			return fmt.Errorf("app.Driver.DoneTaxiCallRequest: error while get taxi call request: %w", err)
		}

		if taxiCallRequest.CurrentState.Complete() {
			// TODO (taeykeom) change error code
			return fmt.Errorf("app.Driver.DoneTaxiCallRequest: already completed call request: %w", value.ErrAlreadyExpiredCallRequest)
		}

		if taxiCallRequest.DriverId.String != driverId {
			return fmt.Errorf("app.Driver.DoneTaxiCallRequest: forbidden access: %w", value.ErrUnAuthorized)
		}

		if err := taxiCallRequest.UpdateState(requestTime, enum.TaxiCallState_DONE); err != nil {
			return fmt.Errorf("app.Driver.DoneTaxiCallRequest: invalid state change:%w", err)
		}

		taxiCallRequest.BasePrice = req.BasePrice
		taxiCallRequest.UpdateTime = requestTime

		if err := d.repository.taxiCallRequest.Update(ctx, i, taxiCallRequest); err != nil {
			return fmt.Errorf("app.Driver.DoneTaxiCallRequest: error while update taxi call request :%w", err)
		}

		driverTaxiCallContext, err := d.repository.taxiCallRequest.GetDriverTaxiCallContext(ctx, i, driverId)
		if err != nil {
			return fmt.Errorf("app.Driver.RejectTaxiCallRequest: error while get taxi call context:%w", err)
		}

		driverTaxiCallContext.CanReceive = true
		if err := d.repository.taxiCallRequest.UpsertDriverTaxiCallContext(ctx, i, driverTaxiCallContext); err != nil {
			return fmt.Errorf("app.Driver.DoneTaxiCallRequest: error while upsert taxi call context: %w", err)
		}

		taxiCallSettlement := entity.DriverTaxiCallSettlement{
			TaxiCallRequestId: taxiCallRequest.Id,
			SettlementDone:    false,
		}
		if err := d.repository.taxiCallRequest.CreateDriverTaxiCallSettlement(ctx, i, taxiCallSettlement); err != nil {
			return fmt.Errorf("app.Driver.DoneTaxiCallRequest: error while create taxi call settlement: %w", err)
		}

		userCmd := command.NewUserTaxiCallNotificationCommand(taxiCallRequest, entity.TaxiCallTicket{}, entity.DriverTaxiCallContext{})
		if err := d.repository.event.BatchCreate(ctx, i, []entity.Event{userCmd}); err != nil {
			return fmt.Errorf("app.Driver.DoneTaxiCallRequest: error while create event: %w", err)
		}

		// TODO (taekyeom) Do payment command

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
