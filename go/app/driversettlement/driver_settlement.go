package driversettlement

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/repository"
	"github.com/uptrace/bun"
)

type driversettlementApp struct {
	app.Transactor

	repository struct {
		settlement repository.DriverSettlementRepository
	}
}

func (d driversettlementApp) GetExpectedDriverSettlement(ctx context.Context, driverId string) (entity.DriverExpectedSettlement, error) {
	var expectedSettlement entity.DriverExpectedSettlement
	err := d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		es, err := d.repository.settlement.GetDriverExpectedSettlement(ctx, i, driverId)
		if err != nil && !errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("app.driversettlementApp.GetExpectedDriverSetttlement: error while select expected settlement: %w", err)
		}
		if errors.Is(err, value.ErrNotFound) {
			expectedSettlement = entity.DriverExpectedSettlement{
				DriverId:       driverId,
				ExpectedAmount: 0,
			}
		}
		expectedSettlement = es

		return nil
	})

	return expectedSettlement, err
}

func (d driversettlementApp) ListDriverSettlementHistory(ctx context.Context, req request.ListDriverSettlementHistoryRequest) ([]entity.DriverSettlementHistory, time.Time, error) {
	var settlementHistories []entity.DriverSettlementHistory
	var pageToken time.Time

	if err := req.Validate(); err != nil {
		return []entity.DriverSettlementHistory{}, time.Time{}, fmt.Errorf("app.driversettlementApp.ListDriverSettlementHistory: error while validate request: %w", err)
	}

	err := d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		histories, newToken, err := d.repository.settlement.ListDriverSettlementHistory(ctx, i, req.DriverId, req.ToPageTokenTime(), req.Count)
		if err != nil {
			return fmt.Errorf("app.driversettlementApp.ListDriverSettlementHistory: error while list driver setttlement history: %w", err)
		}
		settlementHistories = histories
		pageToken = newToken

		return nil
	})

	return settlementHistories, pageToken, err
}
