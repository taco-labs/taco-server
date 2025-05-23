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
	if until := time.Until(taxiProgressCmd.DesiredScheduleTime); until > 0 {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(until):
		}
	}
	return t.process(ctx, event.RetryCount, taxiProgressCmd)
}

func (t taxicallApp) process(ctx context.Context, retryCount int, cmd command.TaxiCallProcessMessage) error {
	return t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		//TODO (taekyeom) make taxi call request failed when retry count exceeded

		taxiCallRequest, err := t.repository.taxiCallRequest.GetById(ctx, i, cmd.TaxiCallRequestId)
		if err != nil {
			return fmt.Errorf("app.taxicall.process [%s]: error while get call request: %w", cmd.TaxiCallRequestId, err)
		}

		// Guard.. commands'state and request's current state must be same
		if string(taxiCallRequest.CurrentState) != cmd.TaxiCallState {
			// TODO (taekyeom) logging late message & add metric for anomaly
			return fmt.Errorf("app.taxicall.process [%s]: taxi call state mismatch %v (actual: %v)",
				cmd.TaxiCallRequestId, cmd.TaxiCallState, taxiCallRequest.CurrentState)
		}

		if taxiCallRequest.CurrentState.Complete() && !cmd.EventTime.Before(taxiCallRequest.UpdateTime) {
			if err := t.repository.taxiCallRequest.DeleteTicketByRequestId(ctx, i, taxiCallRequest.Id); err != nil {
				return fmt.Errorf("app.taxicall.process [%s]: failed to delete ticket: %w", cmd.TaxiCallRequestId, err)
			}
			termiationNotification := command.NewUserTaxiCallNotificationCommand(
				taxiCallRequest,
				entity.TaxiCallTicket{},
				entity.DriverTaxiCallContext{},
			)
			if err := t.repository.event.BatchCreate(ctx, i, []entity.Event{termiationNotification}); err != nil {
				return fmt.Errorf("app.taxicall.process [%s]: failed to create taxi call termination notification: %w", cmd.TaxiCallRequestId, err)
			}
			return nil
		}

		if taxiCallRequest.CurrentState.InDriving() {
			// TODO(taekyeom) Send location push message to user
			return nil
		}

		taxiCallTicket, err := t.repository.taxiCallRequest.GetLatestTicketByRequestId(ctx, i, taxiCallRequest.Id)
		if err != nil && !errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("app.taxicall.process [%s]: error while get call request ticket: %w", cmd.TaxiCallRequestId, err)
		}
		if err != nil && errors.Is(err, value.ErrNotFound) {
			taxiCallTicket = entity.TaxiCallTicket{
				Id:                utils.MustNewUUID(),
				TaxiCallRequestId: taxiCallRequest.Id,
				Attempt:           0,
				AdditionalPrice:   0,
				CreateTime:        cmd.DesiredScheduleTime,
				UpdateTime:        cmd.DesiredScheduleTime,
			}
		}

		if cmd.EventTime.UTC().Before(taxiCallTicket.UpdateTime.UTC()) {
			// TODO (taekyeom) logging late message
			return nil
		}

		validTicketOperation := true
		if !taxiCallTicket.IncreaseAttempt(cmd.DesiredScheduleTime) {
			validTicketOperation = taxiCallTicket.IncreasePrice(taxiCallRequest.RequestMaxAdditionalPrice, cmd.DesiredScheduleTime)
		}

		if !validTicketOperation {
			if err := taxiCallRequest.UpdateState(cmd.DesiredScheduleTime, enum.TaxiCallState_FAILED); err != nil {
				return fmt.Errorf("app.taxicall.process [%s]: failed to update state: %w", cmd.TaxiCallRequestId, err)
			}
			if err := t.repository.taxiCallRequest.Update(ctx, i, taxiCallRequest); err != nil {
				return fmt.Errorf("app.taxicall.process: [%s] failed to update call request to failed state: %w", cmd.TaxiCallRequestId, err)
			}

			taxiCallCmd := command.NewTaxiCallProgressCommand(taxiCallRequest.Id, taxiCallRequest.CurrentState, cmd.DesiredScheduleTime, cmd.DesiredScheduleTime)
			if err := t.repository.event.BatchCreate(ctx, i, []entity.Event{taxiCallCmd}); err != nil {
				return fmt.Errorf("app.taxicall.process [%s]: failed to create taxi call failure event: %w", cmd.TaxiCallRequestId, err)
			}
			return nil
		}

		// Update new ticket
		if err := t.repository.taxiCallRequest.UpsertTicket(ctx, i, taxiCallTicket); err != nil {
			return fmt.Errorf("app.taxicall.process: [%s] error while update new ticket: %w", cmd.TaxiCallRequestId, err)
		}

		// Get drivers
		driverTaxiCallContexts, err := t.repository.taxiCallRequest.
			GetDriverTaxiCallContextWithinRadius(ctx, i, taxiCallRequest.Departure.Point, taxiCallTicket.GetRadius(), taxiCallTicket.Id, cmd.DesiredScheduleTime)
		if err != nil {
			return fmt.Errorf("app.taxicall.process: [%s] error while get driver contexts within radius: %w", cmd.TaxiCallRequestId, err)
		}

		if len(driverTaxiCallContexts) > 0 {
			driverTaxiCallContexts = slices.Map(driverTaxiCallContexts, func(dctx entity.DriverTaxiCallContext) entity.DriverTaxiCallContext {
				dctx.LastReceivedRequestTicket = taxiCallTicket.Id
				dctx.LastReceiveTime = cmd.DesiredScheduleTime
				dctx.RejectedLastRequestTicket = false
				return dctx
			})

			// TODO (taekyeom) pick tit-for-tat drivers
			if err := t.repository.taxiCallRequest.BulkUpsertDriverTaxiCallContext(ctx, i, driverTaxiCallContexts); err != nil {
				return fmt.Errorf("app.taxicall.process: [%s] error while upsert driver contexts within radius: %w", cmd.TaxiCallRequestId, err)
			}
		}

		taxiCallCmd := command.NewTaxiCallProgressCommand(taxiCallRequest.Id, taxiCallRequest.CurrentState,
			cmd.DesiredScheduleTime, cmd.DesiredScheduleTime.Add(time.Second*10))
		userCmd := command.NewUserTaxiCallNotificationCommand(taxiCallRequest, taxiCallTicket, entity.DriverTaxiCallContext{})
		driverCmds := slices.Map(driverTaxiCallContexts, func(i entity.DriverTaxiCallContext) entity.Event {
			return command.NewDriverTaxiCallNotificationCommand(taxiCallRequest, taxiCallTicket, i)
		})
		cmds := append(driverCmds, userCmd, taxiCallCmd)

		if err := t.repository.event.BatchCreate(ctx, i, cmds); err != nil {
			return fmt.Errorf("app.taxicall.process: [%s] error while insert notification command events: %w", cmd.TaxiCallRequestId, err)
		}

		return nil
	})
}
