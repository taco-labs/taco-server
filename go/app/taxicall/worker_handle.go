package taxicall

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/event/command"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/domain/value/enum"
	"github.com/taco-labs/taco/go/utils"
	"github.com/taco-labs/taco/go/utils/slices"
	"github.com/uptrace/bun"
)

func (t taxicallApp) handleEvent(ctx context.Context, event entity.Event) error {
	taxiProgressCmd := command.TaxiCallProcessMessage{}
	err := json.Unmarshal(event.Payload, &taxiProgressCmd)
	if err != nil {
		return fmt.Errorf("app.taxicall.handleEvent: error while unmarshal json: %v", err)
	}

	recieveTime := time.Now()
	if until := time.Until(taxiProgressCmd.DesiredScheduleTime); until > 0 {
		recieveTime = taxiProgressCmd.DesiredScheduleTime
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(until):
		}
	}

	return t.process(ctx, recieveTime, event.RetryCount, taxiProgressCmd)
}

func (t taxicallApp) process(ctx context.Context, receiveTime time.Time, retryCount int, cmd command.TaxiCallProcessMessage) error {
	return t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		taxiCallRequest, err := t.repository.taxiCallRequest.GetById(ctx, i, cmd.TaxiCallRequestId)
		if err != nil {
			return fmt.Errorf("app.taxicall.process [%s]: error while get call request: %w", cmd.TaxiCallRequestId, err)
		}

		//TODO (taekyeom) make taxi call request failed when retry count exceeded
		// if retryCount > 3 {
		// 	// TODO (taekyeom) logging
		// 	if err := taxiCallRequest.UpdateState(receiveTime, enum.TaxiCallState_FAILED); err != nil {
		// 		return fmt.Errorf("app.taxicall.process [%s]: failed to forece mark failed state: %w", cmd.TaxiCallRequestId, err)
		// 	}
		// 	if err := t.repository.taxiCallRequest.Update(ctx, i, taxiCallRequest); err != nil {
		// 		return fmt.Errorf("app.taxicall.process: [%s] failed to forece update failed state: %w", cmd.TaxiCallRequestId, err)
		// 	}

		// 	taxiCallCmd := command.NewTaxiCallProgressCommand(taxiCallRequest.Id, taxiCallRequest.CurrentState, receiveTime, receiveTime)
		// 	if err := t.repository.event.BatchCreate(ctx, i, []entity.Event{taxiCallCmd}); err != nil {
		// 		return fmt.Errorf("app.taxicall.process [%s]: failed to create taxi call force failure event: %w", cmd.TaxiCallRequestId, err)
		// 	}
		// 	return nil
		// }

		// Guard.. commands'state and request's current state must be same
		if string(taxiCallRequest.CurrentState) != cmd.TaxiCallState {
			// TODO (taekyeom) logging late message & add metric for anomaly
			return nil
		}

		var events []entity.Event
		switch taxiCallRequest.CurrentState {
		case enum.TaxiCallState_Requested:
			events, err = t.handleTaxiCallRequested(ctx, cmd.EventTime, receiveTime, taxiCallRequest)
		case enum.TaxiCallState_DRIVER_TO_DEPARTURE:
			events, err = t.handleDriverToDeparture(ctx, cmd.EventTime, receiveTime, taxiCallRequest)
		case enum.TaxiCallState_DRIVER_TO_ARRIVAL:
			events, err = t.handleDriverToArrival(ctx, cmd.EventTime, receiveTime, taxiCallRequest)
		case enum.TaxiCallState_DONE:
			events, err = t.handleDone(ctx, cmd.EventTime, receiveTime, taxiCallRequest)
		case enum.TaxiCallState_USER_CANCELLED:
			events, err = t.handleUserCancelld(ctx, cmd.EventTime, receiveTime, taxiCallRequest)
		case enum.TaxiCallState_DRIVER_CANCELLED:
			events, err = t.handleDriverCancelld(ctx, cmd.EventTime, receiveTime, taxiCallRequest)
		case enum.TaxiCallState_FAILED:
			events, err = t.handleFailed(ctx, cmd.EventTime, receiveTime, taxiCallRequest)
		}

		if err != nil {
			return fmt.Errorf("app.taxicall.process [%s]: failed to handle taxi call progress command: %w", cmd.TaxiCallRequestId, err)
		}

		if len(events) > 0 {
			if err := t.repository.event.BatchCreate(ctx, i, events); err != nil {
				return fmt.Errorf("app.taxicall.process [%s]: failed to insert event: %w", cmd.TaxiCallRequestId, err)
			}
		}

		return nil
	})
}

func (t taxicallApp) handleTaxiCallRequested(ctx context.Context, eventTime time.Time, receiveTime time.Time, taxiCallRequest entity.TaxiCallRequest) ([]entity.Event, error) {
	events := []entity.Event{}
	err := t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		taxiCallTicket, err := t.repository.taxiCallRequest.GetLatestTicketByRequestId(ctx, i, taxiCallRequest.Id)
		if err != nil && !errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("app.taxicall.handleTaxiCallRequested [%s]: error while get call request ticket: %w", taxiCallRequest.Id, err)
		}
		if err != nil && errors.Is(err, value.ErrNotFound) {
			taxiCallTicket = entity.TaxiCallTicket{
				TaxiCallRequestId: taxiCallRequest.Id,
				Attempt:           0,
				AdditionalPrice:   0,
				TicketId:          utils.MustNewUUID(),
				CreateTime:        eventTime,
			}
		}

		if eventTime.Before(taxiCallTicket.CreateTime) {
			// TODO (taekyeom) logging late message
			return nil
		}

		taxiCallTicket, validTicketOperation := taxiCallTicket.Step(taxiCallRequest.RequestMaxAdditionalPrice, receiveTime)

		if !validTicketOperation {
			if err := taxiCallRequest.UpdateState(receiveTime, enum.TaxiCallState_FAILED); err != nil {
				return fmt.Errorf("app.taxicall.handleTaxiCallRequested [%s]: failed to update state: %w", taxiCallRequest.Id, err)
			}
			if err := t.repository.taxiCallRequest.Update(ctx, i, taxiCallRequest); err != nil {
				return fmt.Errorf("app.taxicall.handleTaxiCallRequested: [%s] failed to update call request to failed state: %w", taxiCallRequest.Id, err)
			}

			events = append(events,
				command.NewTaxiCallProgressCommand(taxiCallRequest.Id, taxiCallRequest.CurrentState, receiveTime, receiveTime))
			return nil
		}

		// If ticket exists, return nil (late or duplicated message)
		exists, err := t.repository.taxiCallRequest.TicketExists(ctx, i, taxiCallTicket)
		if err != nil {
			return fmt.Errorf("app.taxicall.handleTaxiCallRequested: [%s] error while check taxi call ticket existance: %w", taxiCallRequest.Id, err)
		}
		if exists {
			// TODO (taekyeom) duplication & late event logging
			return nil
		}

		// Update new ticket
		if err := t.repository.taxiCallRequest.CreateTicket(ctx, i, taxiCallTicket); err != nil {
			return fmt.Errorf("app.taxicall.handleTaxiCallRequested: [%s] error while create new ticket: %w", taxiCallRequest.Id, err)
		}

		// Get drivers
		driverTaxiCallContexts, err := t.repository.taxiCallRequest.
			GetDriverTaxiCallContextWithinRadius(ctx, i, taxiCallRequest.Departure.Point, taxiCallTicket.GetRadius(),
				taxiCallTicket.TaxiCallRequestId, receiveTime)
		if err != nil {
			return fmt.Errorf("app.taxicall.handleTaxiCallRequested: [%s] error while get driver contexts within radius: %w", taxiCallRequest.Id, err)
		}

		if len(driverTaxiCallContexts) > 0 {
			driverTaxiCallContexts = slices.Map(driverTaxiCallContexts, func(dctx entity.DriverTaxiCallContext) entity.DriverTaxiCallContext {
				dctx.LastReceivedRequestTicket = taxiCallTicket.TaxiCallRequestId
				dctx.LastReceiveTime = receiveTime
				dctx.RejectedLastRequestTicket = false
				return dctx
			})

			// TODO (taekyeom) pick tit-for-tat drivers
			if err := t.repository.taxiCallRequest.BulkUpsertDriverTaxiCallContext(ctx, i, driverTaxiCallContexts); err != nil {
				return fmt.Errorf("app.taxicall.handleTaxiCallRequested: [%s] error while upsert driver contexts within radius: %w", taxiCallRequest.Id, err)
			}
		}

		taxiCallCmd := command.NewTaxiCallProgressCommand(taxiCallRequest.Id, taxiCallRequest.CurrentState,
			receiveTime, receiveTime.Add(time.Second*10))
		userCmd := command.NewUserTaxiCallNotificationCommand(taxiCallRequest, taxiCallTicket, entity.DriverTaxiCallContext{})
		driverCmds := slices.Map(driverTaxiCallContexts, func(i entity.DriverTaxiCallContext) entity.Event {
			return command.NewDriverTaxiCallNotificationCommand(taxiCallRequest, taxiCallTicket, i)
		})

		events = append(events, taxiCallCmd)
		events = append(events, userCmd)
		events = append(events, driverCmds...)

		return nil
	})

	if err != nil {
		return []entity.Event{}, err
	}

	return events, nil
}

func (t taxicallApp) handleDriverToDeparture(ctx context.Context, eventTime time.Time, recieveTime time.Time, taxiCallRequest entity.TaxiCallRequest) ([]entity.Event, error) {
	events := []entity.Event{}
	err := t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		// TODO(taekyeom) Send location push message to user
		if eventTime.Before(taxiCallRequest.UpdateTime) {
			return nil
		}

		// TODO(taekyeom) 티켓 수신한 다른 기사분들을 다시 수신 가능한 상태로 만들어야 함

		taxiCallTicket, err := t.repository.taxiCallRequest.GetLatestTicketByRequestId(ctx, i, taxiCallRequest.Id)
		if err != nil && !errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("app.taxicall.handleDriverToDeparture [%s]: error while get call request ticket: %w", taxiCallRequest.Id, err)
		}

		driverTaxiCallContext, err := t.repository.taxiCallRequest.GetDriverTaxiCallContext(ctx, i, taxiCallRequest.DriverId.String)
		if err != nil && !errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("app.taxicall.handleDriverToDeparture [%s]: error while get call request ticket: %w", taxiCallRequest.Id, err)
		}

		events = append(events, command.NewUserTaxiCallNotificationCommand(
			taxiCallRequest,
			taxiCallTicket,
			driverTaxiCallContext,
		))

		return nil
	})

	if err != nil {
		return []entity.Event{}, err
	}

	return events, nil
}

func (t taxicallApp) handleDriverToArrival(ctx context.Context, eventTime time.Time, recieveTime time.Time, taxiCallRequest entity.TaxiCallRequest) ([]entity.Event, error) {
	// TODO(taekyeom) Send location push message to user
	return []entity.Event{}, nil
}

func (t taxicallApp) handleDone(ctx context.Context, eventTime time.Time, recieveTime time.Time, taxiCallRequest entity.TaxiCallRequest) ([]entity.Event, error) {
	// TODO(taekyeom) logging
	if eventTime.Before(taxiCallRequest.UpdateTime) {
		return []entity.Event{}, nil
	}

	events := []entity.Event{}
	err := t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		if err := t.repository.taxiCallRequest.DeleteTicketByRequestId(ctx, i, taxiCallRequest.Id); err != nil {
			return fmt.Errorf("app.taxicall.handleDone [%s]: failed to delete ticket: %w", taxiCallRequest.Id, err)
		}
		events = append(events, command.NewUserTaxiCallNotificationCommand(
			taxiCallRequest,
			entity.TaxiCallTicket{},
			entity.DriverTaxiCallContext{},
		))
		return nil
	})

	if err != nil {
		return []entity.Event{}, err
	}

	return events, nil
}

func (t taxicallApp) handleUserCancelld(ctx context.Context, eventTime time.Time, recieveTime time.Time, taxiCallRequest entity.TaxiCallRequest) ([]entity.Event, error) {
	// TODO(taekyeom) logging
	if eventTime.Before(taxiCallRequest.UpdateTime) {
		return []entity.Event{}, nil
	}

	events := []entity.Event{}
	err := t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		if err := t.repository.taxiCallRequest.DeleteTicketByRequestId(ctx, i, taxiCallRequest.Id); err != nil {
			return fmt.Errorf("app.taxicall.handleUserCancelled [%s]: failed to delete ticket: %w", taxiCallRequest.Id, err)
		}

		if taxiCallRequest.DriverId.Valid {
			events = append(events, command.NewDriverTaxiCallNotificationCommand(
				taxiCallRequest,
				entity.TaxiCallTicket{},
				entity.DriverTaxiCallContext{},
			))
		}

		return nil
	})

	if err != nil {
		return []entity.Event{}, err
	}

	return events, nil
}

func (t taxicallApp) handleDriverCancelld(ctx context.Context, eventTime time.Time, recieveTime time.Time, taxiCallRequest entity.TaxiCallRequest) ([]entity.Event, error) {
	// TODO(taekyeom) logging
	if eventTime.Before(taxiCallRequest.UpdateTime) {
		return []entity.Event{}, nil
	}

	events := []entity.Event{}
	err := t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		if err := t.repository.taxiCallRequest.DeleteTicketByRequestId(ctx, i, taxiCallRequest.Id); err != nil {
			return fmt.Errorf("app.taxicall.handleDriverCancelled [%s]: failed to delete ticket: %w", taxiCallRequest.Id, err)
		}

		events = append(events, command.NewUserTaxiCallNotificationCommand(
			taxiCallRequest,
			entity.TaxiCallTicket{},
			entity.DriverTaxiCallContext{},
		))

		return nil
	})

	if err != nil {
		return []entity.Event{}, err
	}

	return events, nil
}

func (t taxicallApp) handleFailed(ctx context.Context, eventTime time.Time, recieveTime time.Time, taxiCallRequest entity.TaxiCallRequest) ([]entity.Event, error) {
	// TODO(taekyeom) logging
	if eventTime.Before(taxiCallRequest.UpdateTime) {
		return []entity.Event{}, nil
	}

	events := []entity.Event{}
	err := t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		if err := t.repository.taxiCallRequest.DeleteTicketByRequestId(ctx, i, taxiCallRequest.Id); err != nil {
			return fmt.Errorf("app.taxicall.handleFailed [%s]: failed to delete ticket: %w", taxiCallRequest.Id, err)
		}
		events = append(events, command.NewUserTaxiCallNotificationCommand(
			taxiCallRequest,
			entity.TaxiCallTicket{},
			entity.DriverTaxiCallContext{},
		))
		return nil
	})

	if err != nil {
		return []entity.Event{}, err
	}

	return events, nil
}
