package driversettlement

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/event/command"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/domain/value/enum"
	"github.com/taco-labs/taco/go/utils"
	"github.com/uptrace/bun"
)

func (d driversettlementApp) Accept(ctx context.Context, event entity.Event) bool {
	return strings.HasPrefix(event.EventUri, command.EventUri_DriverSettlementPrefix)
}

func (d driversettlementApp) OnFailure(ctx context.Context, event entity.Event, lastErr error) error {
	var err error
	switch event.EventUri {
	case command.EventUri_DriverSettlementTransferRequest:
		err = d.handleTransferRequestFailure(ctx, event, lastErr)
	case command.EventUri_DriverSettlementTransferExecution:
		err = d.handleTransferExecutionFailure(ctx, event, lastErr)
	}
	return err
}

func (d driversettlementApp) Process(ctx context.Context, event entity.Event) error {
	requestTime := time.Now()
	defer func() {
		tags := []string{"eventUri", event.EventUri}
		now := time.Now()
		d.service.metric.Timing("WorkerProcessTime", now.Sub(requestTime), tags...)
	}()

	select {
	case <-ctx.Done():
		return nil
	default:
		var err error
		switch event.EventUri {
		case command.EventUri_DriverSettlementRequest:
			err = d.handleSettlementRequest(ctx, event)
		case command.EventUri_DriverSettlementTransferRequest:
			err = d.handleSettlementTransferRequest(ctx, event)
		case command.EventUri_DriverSettlementTransferExecution:
			err = d.handleSettlementTransferExecution(ctx, event)
		case command.EventUri_DriverSettlementTransferSuccess:
			err = d.handleSettlementTransferSuccess(ctx, event)
		case command.EventUri_DriverSettlementTransferFail:
			err = d.handleSettlementTransferFail(ctx, event)
		default:
			return fmt.Errorf("invalid event uri '%s': %w", event.EventUri, value.ErrInvalidOperation)
		}
		return err
	}
}

func (d driversettlementApp) handleSettlementRequest(ctx context.Context, event entity.Event) error {
	cmd := command.DriverSettlementRequestCommand{}
	err := json.Unmarshal(event.Payload, &cmd)
	if err != nil {
		return fmt.Errorf("app.driversettlementApp.handleSettlementRequest: error while unmarshal command: %w", err)
	}

	ctx = utils.SetRequestTime(ctx, cmd.RequestTime)

	promotionReward, err := d.ApplyDriverSettlementPromotionReward(
		ctx,
		request.ApplyDriverSettlementPromotionRewardRequest{
			DriverId: cmd.DriverId,
			OrderId:  cmd.TaxiCallRequestId,
		},
	)
	if err != nil {
		return fmt.Errorf("app.driversettlementApp.handleSettlementRequest: error while applying promotion reward: %w", err)
	}

	settlementAmount := cmd.Amount + promotionReward

	err = d.ApplyDriverSettlementRequest(ctx, cmd.DriverId, cmd.TaxiCallRequestId, settlementAmount)
	if err != nil {
		return fmt.Errorf("app.driversettlementApp.handleSettlementRequest: error while apply driver settlement: %w", err)
	}

	return nil
}

func (d driversettlementApp) handleSettlementTransferRequest(ctx context.Context, event entity.Event) error {
	cmd := command.DriverSettlementTransferRequestCommand{}
	err := json.Unmarshal(event.Payload, &cmd)
	if err != nil {
		return fmt.Errorf("app.driversettlementApp.handleSettlementTransferRequest: error while unmarshal command: %w", err)
	}

	var inflightRequest entity.DriverInflightSettlementTransfer

	err = d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		req, err := d.repository.settlement.GetInflightSettlementTransferByDriverId(ctx, i, cmd.DriverId)
		if errors.Is(err, value.ErrNotFound) {
			return nil
		}
		if err != nil && errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("app.driversettlementApp.handleSettlementTransferRequest: error while get inflight transfer: %w", err)
		}

		if req.State != enum.SettlementTransferProcessState_Received {
			return nil
		}

		inflightRequest = req

		return nil
	})

	if err != nil {
		return err
	}

	settlementTransferRequest := value.SettlementTransferRequest{
		DriverId:          inflightRequest.DriverId,
		TransferKey:       inflightRequest.TransferId,
		BankTransactionId: inflightRequest.BankTransactionId,
		Amount:            inflightRequest.AmountWithoutTax,
		Message:           inflightRequest.Message,
	}

	transferRequest, err := d.service.settlementAccount.TransferRequest(ctx, settlementTransferRequest)
	// TODO (taekyeom) already exist error handle
	if err != nil {
		return fmt.Errorf("app.driversettlementApp.handleSettlementTransferRequest: error while request transfer: %w", err)
	}

	return d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		inflightRequest.State = enum.SettlementTransferProcessState_REQUESTED
		inflightRequest.ExecutionKey = transferRequest.ExecutionKey
		inflightRequest.UpdateTime = event.CreateTime

		if err := d.repository.settlement.UpdateInflightSettlementTransfer(ctx, i, inflightRequest); err != nil {
			return fmt.Errorf("app.driversettlementApp.handleSettlementTransferRequest: error while update inflight request: %w", err)
		}

		cmd := command.NewDriverSettlementTransferExecutionCommand(inflightRequest.DriverId)
		if err := d.repository.event.BatchCreate(ctx, i, []entity.Event{cmd}); err != nil {
			return fmt.Errorf("app.driversettlementApp.handleSettlementTransferRequest: error while create command: %w", err)
		}

		return nil
	})
}

func (d driversettlementApp) handleSettlementTransferExecution(ctx context.Context, event entity.Event) error {
	cmd := command.DriverSettlementTransferExecutionCommand{}
	err := json.Unmarshal(event.Payload, &cmd)
	if err != nil {
		return fmt.Errorf("app.driversettlementApp.handleSettlementTransferExecution: error while unmarshal command: %w", err)
	}

	var inflightRequest entity.DriverInflightSettlementTransfer

	err = d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		req, err := d.repository.settlement.GetInflightSettlementTransferByDriverId(ctx, i, cmd.DriverId)
		if errors.Is(err, value.ErrNotFound) {
			return nil
		}
		if err != nil && errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("app.driversettlementApp.handleSettlementTransferExecution: error while get inflight transfer: %w", err)
		}

		if req.State != enum.SettlementTransferProcessState_REQUESTED {
			return nil
		}

		inflightRequest = req

		return nil
	})

	if err != nil {
		return fmt.Errorf("app.driversettlementApp.handleSettlementTransferExecution: error while update state: %w", err)
	}

	settlementTransferExecution := value.SettlementTransfer{ExecutionKey: inflightRequest.ExecutionKey}
	if err := d.service.settlementAccount.TransferExecution(ctx, settlementTransferExecution); err != nil {
		return fmt.Errorf("app.driversettlementApp.handleSettlementTransferExecution: error while execute transfer: %w", err)
	}

	return d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		inflightRequest.State = enum.SettlementTransferProcessState_EXECUTED
		inflightRequest.UpdateTime = event.CreateTime

		if err := d.repository.settlement.UpdateInflightSettlementTransfer(ctx, i, inflightRequest); err != nil {
			return fmt.Errorf("app.driversettlementApp.handleSettlementTransferExecution: error while update inflight request: %w", err)
		}

		return nil
	})
}

func (d driversettlementApp) handleSettlementTransferSuccess(ctx context.Context, event entity.Event) error {
	cmd := command.DriverSettlementTransferSuccessCommand{}
	err := json.Unmarshal(event.Payload, &cmd)
	if err != nil {
		return fmt.Errorf("app.driversettlementApp.handleSettlementTransferSuccess: error while unmarshal command: %w", err)
	}

	return d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		inflightRequest, err := d.repository.settlement.GetInflightSettlementTransferByDriverId(ctx, i, cmd.DriverId)
		if errors.Is(err, value.ErrNotFound) {
			return nil
		}
		if err != nil && errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("app.driversettlementApp.handleSettlementTransferSuccess: error while get inflight transfer: %w", err)
		}

		if inflightRequest.State != enum.SettlementTransferProcessState_EXECUTED {
			return nil
		}

		// 1. Subtract total settlement
		if err := d.repository.settlement.UpdateTotalDriverSettlement(ctx, i, inflightRequest.DriverId, -inflightRequest.Amount); err != nil {
			return fmt.Errorf("app.driversettlementApp.handleSettlementTransferSuccess: error while update total settlement: %w", err)
		}

		// 2. Remove settlement request
		if err := d.repository.settlement.BatchDeleteDriverSettlementRequest(ctx, i, inflightRequest.DriverId, inflightRequest.CreateTime); err != nil {
			return fmt.Errorf("app.driversettlementApp.handleSettlementTransferSuccess: error while delete settlement requests: %w", err)
		}

		if err := d.repository.settlement.DeleteInflightSettlementTransfer(ctx, i, inflightRequest); err != nil {
			return fmt.Errorf("app.driversettlementApp.handleSettlementTransferSuccess: error while delete failed inflight transfer: %w", err)
		}

		// 3. Add history
		settlementHistory := entity.DriverSettlementHistory{
			DriverId:         cmd.DriverId,
			Amount:           inflightRequest.Amount,
			AmountWithoutTax: inflightRequest.AmountWithoutTax,
			Bank:             cmd.Bank,
			AccountNumber:    cmd.AccountNumber,
			RequestTime:      inflightRequest.CreateTime,
			CreateTime:       event.CreateTime,
		}

		if err := d.repository.settlement.CreateDriverSettlementHistory(ctx, i, settlementHistory); err != nil {
			return fmt.Errorf("app.driversettlementApp.handleSettlementTransferSuccess: error while create settlement history: %w", err)
		}

		// 4. Notify
		cmd := settlementTransferSuccessMessage(inflightRequest.DriverId, inflightRequest.Amount)
		if err := d.repository.event.BatchCreate(ctx, i, []entity.Event{cmd}); err != nil {
			return fmt.Errorf("app.driversettlementApp.handleSettlementTransferSuccess: error while create push notification command: %w", err)
		}
		return nil
	})
}

func (d driversettlementApp) handleSettlementTransferFail(ctx context.Context, event entity.Event) error {
	cmd := command.DriverSettlementTransferFailCommand{}
	err := json.Unmarshal(event.Payload, &cmd)
	if err != nil {
		return fmt.Errorf("app.driversettlementApp.handleSettlementTransferFail: error while unmarshal command: %w", err)
	}

	return d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		inflightRequest, err := d.repository.settlement.GetInflightSettlementTransferByDriverId(ctx, i, cmd.DriverId)
		if errors.Is(err, value.ErrNotFound) {
			return nil
		}
		if err != nil && errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("app.driversettlementApp.handleSettlementTransferFail: error while get inflight transfer: %w", err)
		}

		if inflightRequest.State != enum.SettlementTransferProcessState_EXECUTED {
			return nil
		}

		failedRequest := entity.DriverFailedSettlementTransfer{
			TransferId:        inflightRequest.TransferId,
			DriverId:          inflightRequest.DriverId,
			ExecutionKey:      inflightRequest.ExecutionKey,
			BankTransactionId: inflightRequest.BankTransactionId,
			Amount:            inflightRequest.Amount,
			AmountWithoutTax:  inflightRequest.AmountWithoutTax,
			Message:           inflightRequest.Message,
			FailureMessage:    cmd.FailureMessage,
			CreateTime:        event.CreateTime,
		}

		if err := d.repository.settlement.CreateFailedSettlementTransfer(ctx, i, failedRequest); err != nil {
			return fmt.Errorf("app.driversettlementApp.handleSettlementTransferFail: error while create failed transfer: %w", err)
		}

		if err := d.repository.settlement.DeleteInflightSettlementTransfer(ctx, i, inflightRequest); err != nil {
			return fmt.Errorf("app.driversettlementApp.handleSettlementTransferFail: error while delete failed inflight transfer: %w", err)
		}

		cmd := settlementTransferFailureMessage(inflightRequest.DriverId)
		if err := d.repository.event.BatchCreate(ctx, i, []entity.Event{cmd}); err != nil {
			return fmt.Errorf("app.driversettlementApp.handleSettlementTransferFail: error while create push notification command: %w", err)
		}
		return nil
	})
}

func (d driversettlementApp) handleTransferRequestFailure(ctx context.Context, event entity.Event, lastErr error) error {
	cmd := command.DriverSettlementTransferRequestCommand{}
	err := json.Unmarshal(event.Payload, &cmd)
	if err != nil {
		return fmt.Errorf("app.driversettlementApp.handleSettlementTransferRequest: error while unmarshal command: %w", err)
	}

	return d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		cmd := command.NewDriverSettlementTransferFailCommand(cmd.DriverId, lastErr.Error())
		if err := d.repository.event.BatchCreate(ctx, i, []entity.Event{cmd}); err != nil {
			return fmt.Errorf("app.driversettlementApp.handleTransferRequestFailure: error while create failure command: %w", err)
		}
		return nil
	})
}

func (d driversettlementApp) handleTransferExecutionFailure(ctx context.Context, event entity.Event, lastErr error) error {
	cmd := command.DriverSettlementTransferExecutionCommand{}
	err := json.Unmarshal(event.Payload, &cmd)
	if err != nil {
		return fmt.Errorf("app.driversettlementApp.handleSettlementTransferExecution: error while unmarshal command: %w", err)
	}

	return d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		cmd := command.NewDriverSettlementTransferFailCommand(cmd.DriverId, lastErr.Error())
		if err := d.repository.event.BatchCreate(ctx, i, []entity.Event{cmd}); err != nil {
			return fmt.Errorf("app.driversettlementApp.handleTransferExecutionFailure: error while create failure command: %w", err)
		}
		return nil
	})
}
