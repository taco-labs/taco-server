package driversettlement

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/event/command"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/utils"
	"github.com/uptrace/bun"
	"go.uber.org/zap"
)

func (d driversettlementApp) Accept(ctx context.Context, event entity.Event) bool {
	return strings.HasPrefix(event.EventUri, command.EventUri_DriverSettlementPrefix)
}

func (d driversettlementApp) Process(ctx context.Context, event entity.Event) error {
	select {
	case <-ctx.Done():
		return nil
	default:
		var err error
		switch event.EventUri {
		case command.EventUri_DriverSettlementRequest:
			err = d.handleSettlementRequest(ctx, event)
		case command.EventUri_DriverSettlementDone:
			err = d.handleSettlementDone(ctx, event)
		}
		return err
	}
}

func (d driversettlementApp) handleSettlementRequest(ctx context.Context, event entity.Event) error {
	cmd := command.DriverSettlementRequestCommand{}
	logger := utils.GetLogger(ctx)
	err := json.Unmarshal(event.Payload, &cmd)
	if err != nil {
		return fmt.Errorf("app.driversettlementApp.handleSettlementRequest: error while unmarshal command: %w", err)
	}

	return d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		settlementRequest, err := d.repository.settlement.GetDriverSettlementRequest(ctx, i, cmd.TaxiCallRequestId)
		if err != nil && !errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("app.driversettlementApp.handleSettlementRequest: error while get settlement request: %w", err)
		}
		if settlementRequest.TaxiCallRequestId != "" {
			logger.Warn("duplicated message",
				zap.Any("cmd", cmd),
				zap.String("type", "settlement"),
				zap.String("method", "handleSettlementRequest"),
			)
			return nil
		}

		settlementRequest = entity.DriverSettlementRequest{
			TaxiCallRequestId: cmd.TaxiCallRequestId,
			DriverId:          cmd.DriverId,
			Amount:            cmd.Amount,
			CreateTime:        cmd.RequestTime,
		}
		if err := d.repository.settlement.CreateDriverSettlementRequest(ctx, i, settlementRequest); err != nil {
			return fmt.Errorf("app.driversettlementApp.handleSettlementRequest: error while create settlement request: %w", err)
		}
		if err := d.repository.settlement.UpdateExpectedDriverSettlement(ctx, i, cmd.DriverId, cmd.Amount); err != nil {
			return fmt.Errorf("app.driversettlementApp.handleSettlementRequest: error while update expected settlement amount: %w", err)
		}

		return nil
	})
}

func (d driversettlementApp) handleSettlementDone(ctx context.Context, event entity.Event) error {
	cmd := command.DriverSettlementDoneCommand{}
	logger := utils.GetLogger(ctx)
	err := json.Unmarshal(event.Payload, &cmd)
	if err != nil {
		return fmt.Errorf("app.driversettlementApp.handleSettlementDone: error while unmarshal command: %w", err)
	}

	return d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		settlementHistory, err := d.repository.settlement.GetDriverSettlementHistory(ctx, i, cmd.DriverId, cmd.SettlementPeriodStart, cmd.SettlementPeriodEnd)
		if err != nil && !errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("app.driversettlementApp.handleSettlementDone: error while get settlement history: %w", err)
		}
		if settlementHistory.DriverId != "" {
			logger.Warn("duplicated message",
				zap.Any("cmd", cmd),
				zap.String("type", "settlement"),
				zap.String("method", "handleSettlementDone"),
			)
			return nil
		}

		settlementHistory = entity.DriverSettlementHistory{
			DriverId:              cmd.DriverId,
			SettlementPeriodStart: cmd.SettlementPeriodStart,
			SettlementPeriodEnd:   cmd.SettlementPeriodEnd,
			CreateTime:            event.CreateTime,
			Amount:                cmd.Amount,
		}
		if err := d.repository.settlement.CreateDriverSettlementHistory(ctx, i, settlementHistory); err != nil {
			return fmt.Errorf("app.driversettlementApp.handleSettlementDone: error while create settlement history: %w", err)
		}
		// TODO (taekyeom) 정산 완료된 request entity 처리 필요
		if err := d.repository.settlement.UpdateExpectedDriverSettlement(ctx, i, cmd.DriverId, -cmd.Amount); err != nil {
			return fmt.Errorf("app.driversettlementApp.handleSettlementDone: error while update expected settlement amount: %w", err)
		}

		return nil
	})
}
