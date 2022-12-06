package taxicall

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/event/command"
	"github.com/taco-labs/taco/go/domain/value/enum"
	"github.com/uptrace/bun"
)

func (t taxicallApp) Accept(ctx context.Context, event entity.Event) bool {
	return strings.HasPrefix(event.EventUri, command.EventUri_TaxiCallPrefix)
}

func (t taxicallApp) OnFailure(ctx context.Context, event entity.Event, lastErr error) error {
	switch event.EventUri {
	case command.EventUri_TaxiCallProcess:
		return t.handleTaxiCallProgressFailure(ctx, event, lastErr)
	default:
		return nil
	}
}

func (t taxicallApp) Process(ctx context.Context, event entity.Event) error {
	select {
	case <-ctx.Done():
		return nil
	default:
		return t.handleEvent(ctx, event)
	}
}

func (t taxicallApp) handleTaxiCallProgressFailure(ctx context.Context, event entity.Event, lastErr error) error {
	cmd := command.TaxiCallProcessMessage{}
	err := json.Unmarshal(event.Payload, &cmd)
	if err != nil {
		return fmt.Errorf("app.taxicall.handleEvent: error while unmarshal json: %v", err)
	}

	receiveTime := event.CreateTime
	return t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		taxiCallRequest, err := t.repository.taxiCallRequest.GetById(ctx, i, cmd.TaxiCallRequestId)
		if err != nil {
			return fmt.Errorf("app.taxicall.process [%s]: error while get call request: %w", cmd.TaxiCallRequestId, err)
		}

		if err := taxiCallRequest.UpdateState(receiveTime, enum.TaxiCallState_FAILED); err != nil {
			return fmt.Errorf("app.taxicall.process [%s]: failed to forece mark failed state: %w", cmd.TaxiCallRequestId, err)
		}
		if err := t.repository.taxiCallRequest.Update(ctx, i, taxiCallRequest); err != nil {
			return fmt.Errorf("app.taxicall.process: [%s] failed to forece update failed state: %w", cmd.TaxiCallRequestId, err)
		}

		taxiCallCmd := command.NewTaxiCallProgressCommand(taxiCallRequest.Id, taxiCallRequest.CurrentState, receiveTime, receiveTime)
		if err := t.repository.event.BatchCreate(ctx, i, []entity.Event{taxiCallCmd}); err != nil {
			return fmt.Errorf("app.taxicall.process [%s]: failed to create taxi call force failure event: %w", cmd.TaxiCallRequestId, err)
		}
		return nil
	})
}
