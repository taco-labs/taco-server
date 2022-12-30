package driversettlement

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/event/command"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/domain/value/analytics"
	"github.com/taco-labs/taco/go/domain/value/enum"
	"github.com/taco-labs/taco/go/repository"
	"github.com/taco-labs/taco/go/service"
	"github.com/taco-labs/taco/go/utils"
	"github.com/uptrace/bun"
)

type driversettlementApp struct {
	app.Transactor

	repository struct {
		settlement repository.DriverSettlementRepository
		event      repository.EventRepository
		analytics  repository.AnalyticsRepository
	}

	service struct {
		settlementAccount service.SettlementAccountService
	}
}

// TODO (taekyeom) To be parameterized
func getSettlementRequestableTime(t time.Time) time.Time {
	loc, _ := time.LoadLocation("Asia/Seoul")

	timeInLocation := t.In(loc)

	return time.Date(timeInLocation.Year(), timeInLocation.Month(), timeInLocation.Day(), 0, 0, 0, 0, timeInLocation.Location()).
		AddDate(0, 0, -14).In(time.UTC)
}

func (d driversettlementApp) GetExpectedDriverSettlement(ctx context.Context, driverId string) (entity.DriverTotalSettlement, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	settlementRequestableTime := getSettlementRequestableTime(requestTime)

	var expectedSettlement entity.DriverTotalSettlement

	err := d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		es, err := d.repository.settlement.GetDriverTotalSettlement(ctx, i, driverId)
		if errors.Is(err, value.ErrNotFound) {
			expectedSettlement = entity.DriverTotalSettlement{
				DriverId:    driverId,
				TotalAmount: 0,
			}
		}
		if err != nil {
			return fmt.Errorf("app.driversettlementApp.GetExpectedDriverSetttlement: error while select expected settlement: %w", err)
		}
		expectedSettlement = es

		promotionAmount, err := d.repository.settlement.GetDriverPromotionSettlementReward(ctx, i, driverId)
		if errors.Is(err, value.ErrNotFound) {
			promotionAmount = entity.DriverPromotionSettlementReward{
				DriverId:    driverId,
				TotalAmount: 0,
			}
		}
		if err != nil {
			return fmt.Errorf("app.driversettlementApp.GetExpectedDriverSetttlement: error while get promotion settlement: %w", err)
		}
		expectedSettlement.TotalAmount += promotionAmount.TotalAmount

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

func (d driversettlementApp) RequestSettlementTransfer(ctx context.Context, settlementAccount entity.DriverSettlementAccount) (int, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)
	settlementRequestableTime := getSettlementRequestableTime(requestTime)
	var expectedTransferAmount int

	err := d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		// Check inflight settlement transfer exists
		inflightSettlementTransfer, err := d.repository.settlement.GetInflightSettlementTransferByDriverId(ctx, i, settlementAccount.DriverId)
		if err != nil && !errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("app.driversettlementApp.RequestSettlementTransfer: error while get inflight transfer: %w", err)
		}
		if inflightSettlementTransfer.TransferId != "" {
			return fmt.Errorf("app.driversettlementApp.RequestSettlementTransfer: already inflight settlement transfer exists: %w", value.ErrAlreadyExists)
		}

		transferAmount, err := d.repository.settlement.AggregateDriverRequestableSettlement(ctx, i, settlementAccount.DriverId, settlementRequestableTime)
		if err != nil {
			return fmt.Errorf("app.driversettlementApp.RequestSettlementTransfer: error while get transferable amount: %w", err)
		}

		if transferAmount == 0 {
			return fmt.Errorf("app.driversettlementApp.RequestSettlementTransfer: no transferable amount: %w", value.ErrInvalidOperation)
		}

		inflightSettlementTransfer = entity.DriverInflightSettlementTransfer{
			TransferId:        utils.MustNewUUID(),
			DriverId:          settlementAccount.DriverId,
			BankTransactionId: settlementAccount.BankTransactionId,
			Amount:            transferAmount,
			AmountWithoutTax:  entity.ExpectedSettlementAmountWithoutTax(transferAmount),
			Message:           "타코 정산", // TODO (taekyeom) more precise message?
			State:             enum.SettlementTransferProcessState_Received,
			CreateTime:        requestTime,
			UpdateTime:        requestTime,
		}

		if err := d.repository.settlement.CreateInflightSettlementTransfer(ctx, i, inflightSettlementTransfer); err != nil {
			return fmt.Errorf("app.driversettlementApp.RequestSettlementTransfer: error while create inflight transfer request: %w", err)
		}

		cmd := command.NewDriverSettlementTransferRequestCommand(inflightSettlementTransfer.DriverId)
		if err := d.repository.event.BatchCreate(ctx, i, []entity.Event{cmd}); err != nil {
			return fmt.Errorf("app.driversettlementApp.RequestSettlementTransfer: error while create command: %w", err)
		}

		expectedTransferAmount = transferAmount

		return nil
	})

	if err != nil {
		return 0, err
	}

	return expectedTransferAmount, nil
}

func (d driversettlementApp) DriverSettlementTransferSuccessCallback(ctx context.Context, req request.DriverSettlementTransferSuccessCallbackRequest) error {
	return d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		cmd := command.NewDriverSettlementTransferSuccessCommand(req.DriverId, req.Bank, req.AccountNumber)
		if err := d.repository.event.BatchCreate(ctx, i, []entity.Event{cmd}); err != nil {
			return fmt.Errorf("app.driversettlementApp.DriverSettlementTransferSuccessCallback: error while create success command: %w", err)
		}
		return nil
	})
}

func (d driversettlementApp) DriverSettlementTransferFailureCallback(ctx context.Context, req request.DriverSettlementTransferFailureCallbackRequest) error {
	return d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		cmd := command.NewDriverSettlementTransferFailCommand(req.DriverId, req.FailureMessage)
		if err := d.repository.event.BatchCreate(ctx, i, []entity.Event{cmd}); err != nil {
			return fmt.Errorf("app.driversettlementApp.DriverSettlementTransferFailureCallback: error while create fail command: %w", err)
		}
		return nil
	})
}

func (d driversettlementApp) ApplyDriverSettlementPromotionReward(ctx context.Context, req request.ApplyDriverSettlementPromotionRewardRequest) (int, error) {
	var rewardAmount int
	// TODO (taekyeom) user request가 아닌 형태의 요청은 시간을 undeterministic 하게 받음.. evetn loop에서도 적당히 작업이 필요할 듯 싶음.
	requestTime := utils.GetRequestTimeOrNow(ctx)
	err := d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		promotionReward, err := d.repository.settlement.GetDriverPromotionSettlementReward(ctx, i, req.DriverId)
		if errors.Is(err, value.ErrNotFound) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("app.driversettlementApp.ApplyDriverSettlementPromotionReward: error while get driver promotion reward: %w", err)
		}

		rewardAmount = promotionReward.Apply(req.Amount, req.RewardRate)

		if err := d.repository.settlement.UpdateDriverPromotionSettlementReward(ctx, i, promotionReward); err != nil {
			return fmt.Errorf("app.driversettlementApp.ApplyDriverSettlementPromotionReward: error while update driver promotion reward: %w", err)
		}

		rewardApplyAnalytics := entity.NewAnalytics(requestTime, analytics.DriverPromotionReward{
			DriverId:             req.DriverId,
			TaxiCallRequestId:    req.OrderId,
			RewardRate:           req.RewardRate,
			AfterPromotionAmount: promotionReward.TotalAmount,
			RewardAmount:         rewardAmount,
		})
		if err := d.repository.analytics.Create(ctx, i, rewardApplyAnalytics); err != nil {
			return fmt.Errorf("app.driversettlementApp.ApplyDriverSettlementPromotionReward: error while create analytics: %w", err)
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return rewardAmount, nil
}

func (d driversettlementApp) ReceiveDriverPromotionReward(ctx context.Context, driverId string, receiveTime time.Time) error {
	return d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		driverPromotionReward, err := d.repository.settlement.GetDriverPromotionSettlementReward(ctx, i, driverId)
		if err != nil && !errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("app.driversettlementApp.GiveDriverPromotionReward: error while get driver promotion settlement reward: %w", err)
		}
		if errors.Is(err, value.ErrNotFound) {
			driverPromotionReward = entity.DriverPromotionSettlementReward{
				DriverId:    driverId,
				TotalAmount: 0,
			}
			if err := d.repository.settlement.CreateDriverPromotionSettlementReward(ctx, i, driverPromotionReward); err != nil {
				return fmt.Errorf("app.driversettlementApp.GiveDriverPromotionReward: error while create driver promotion settlement reward: %w", err)
			}
		}

		promotionRewardLimit, err := d.repository.settlement.GetDriverPromotionRewardLimit(ctx, i, driverId)
		if err != nil && !errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("app.driversettlementApp.GiveDriverPromotionReward: error while get driver promotion reward limit: %w", err)
		}
		if errors.Is(err, value.ErrNotFound) {
			promotionRewardLimit = entity.NewDriverPromotionRewardLimit(driverId)
			if err := d.repository.settlement.CreateDriverPromotionRewardLimit(ctx, i, promotionRewardLimit); err != nil {
				return fmt.Errorf("app.driversettlementApp.GiveDriverPromotionReward: error while create driver promotion reward limit: %w", err)
			}
		}

		driverPromotionRewardHistory := entity.NewDriverPromotionRewardHistory(driverId, receiveTime)
		if !driverPromotionRewardHistory.PromotionValid() {
			return nil
		}

		historyExists, err := d.repository.settlement.DriverPromotionRewardHistoryExists(ctx, i, driverPromotionRewardHistory)
		if err != nil {
			return fmt.Errorf("app.driversettlementApp.GiveDriverPromotionReward: error while check driver reward history existance check: %w", err)
		}
		if historyExists {
			return nil
		}

		if !promotionRewardLimit.Receive() {
			return nil
		}

		driverPromotionReward.TotalAmount += entity.DriverRewardReceiveAmount
		if err := d.repository.settlement.UpdateDriverPromotionSettlementReward(ctx, i, driverPromotionReward); err != nil {
			return fmt.Errorf("app.driversettlementApp.GiveDriverPromotionReward: error while update driver promotion settlement reward: %w", err)
		}

		if err := d.repository.settlement.UpdateDriverPromotionRewardLimit(ctx, i, promotionRewardLimit); err != nil {
			return fmt.Errorf("app.driversettlementApp.GiveDriverPromotionReward: error while update driver promotion reward limit: %w", err)
		}

		if err := d.repository.settlement.CreateDriverPromotionRewardHistory(ctx, i, driverPromotionRewardHistory); err != nil {
			return fmt.Errorf("app.driversettlementApp.GiveDriverPromotionReward: error while create driver promotion reward history: %w", err)
		}

		return nil
	})
}
