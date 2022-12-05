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
	"github.com/taco-labs/taco/go/utils"
	"github.com/uptrace/bun"
)

type driversettlementApp struct {
	app.Transactor

	repository struct {
		settlement repository.DriverSettlementRepository
	}
}

// TODO (taekyeom) To be parameterized
func getRequestableTime(t time.Time) time.Time {
	loc, _ := time.LoadLocation("Asia/Seoul")

	timeInLocation := t.In(loc)

	return time.Date(timeInLocation.Year(), timeInLocation.Month(), timeInLocation.Day(), 0, 0, 0, 0, timeInLocation.Location()).
		AddDate(0, 0, -14).In(time.UTC)
}

func (d driversettlementApp) GetExpectedDriverSettlement(ctx context.Context, driverId string) (entity.DriverTotalSettlement, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	settlementRequestableTime := getRequestableTime(requestTime)

	var expectedSettlement entity.DriverTotalSettlement

	err := d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		es, err := d.repository.settlement.GetDriverTotalSettlement(ctx, i, driverId)
		if err != nil && !errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("app.driversettlementApp.GetExpectedDriverSetttlement: error while select expected settlement: %w", err)
		}
		if errors.Is(err, value.ErrNotFound) {
			expectedSettlement = entity.DriverTotalSettlement{
				DriverId:    driverId,
				TotalAmount: 0,
			}
			return nil
		}
		expectedSettlement = es

		requestableAmount, err := d.repository.settlement.AggregateDriverRequestableSettlement(ctx, i, driverId, settlementRequestableTime)
		if err != nil {
			return fmt.Errorf("app.driversettlementApp.GetExpectedDriverSetttlement: error while aggregate requestable settlement: %w", err)
		}
		expectedSettlement.RequestableAmount = requestableAmount

		return nil
	})

	if err != nil {
		return entity.DriverTotalSettlement{}, err
	}

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
