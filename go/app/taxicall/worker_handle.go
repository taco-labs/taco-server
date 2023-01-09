package taxicall

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/event/command"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/domain/value/analytics"
	"github.com/taco-labs/taco/go/domain/value/enum"
	"github.com/taco-labs/taco/go/utils"
	"github.com/taco-labs/taco/go/utils/slices"
	"github.com/uptrace/bun"
	"go.uber.org/zap"
)

func (t taxicallApp) handleEvent(ctx context.Context, event entity.Event) error {
	taxiProgressCmd := command.TaxiCallProcessMessage{}
	err := json.Unmarshal(event.Payload, &taxiProgressCmd)
	if err != nil {
		return fmt.Errorf("app.taxicall.handleEvent: error while unmarshal json: %v", err)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case receiveTime := <-time.After(time.Until(taxiProgressCmd.DesiredScheduleTime)):
		return t.process(ctx, receiveTime, event.Attempt, taxiProgressCmd)
	}
}

func (t taxicallApp) process(ctx context.Context, receiveTime time.Time, retryCount int, cmd command.TaxiCallProcessMessage) error {
	return t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		taxiCallRequest, err := t.repository.taxiCallRequest.GetById(ctx, i, cmd.TaxiCallRequestId)
		if err != nil {
			return fmt.Errorf("app.taxicall.process [%s]: error while get call request: %w", cmd.TaxiCallRequestId, err)
		}

		tags, err := slices.MapErr(taxiCallRequest.TagIds, value.GetTagById)
		if err != nil {
			return fmt.Errorf("app.taxicall.process [%s]: error while get tags: %w", cmd.TaxiCallRequestId, err)
		}
		taxiCallRequest.Tags = tags

		// Guard.. commands'state and request's current state must be same
		if string(taxiCallRequest.CurrentState) != cmd.TaxiCallState {
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
			events, err = t.handleUserCancelled(ctx, cmd.EventTime, receiveTime, taxiCallRequest)
		case enum.TaxiCallState_DRIVER_CANCELLED:
			events, err = t.handleDriverCancelled(ctx, cmd.EventTime, receiveTime, taxiCallRequest)
		case enum.TaxiCallState_FAILED:
			events, err = t.handleFailed(ctx, cmd.EventTime, receiveTime, taxiCallRequest)
		case enum.TaxiCallState_DRIVER_NOT_AVAILABLE:
			events, err = t.handleDriverNotAvailable(ctx, cmd.EventTime, receiveTime, taxiCallRequest)
		}

		if err != nil {
			return fmt.Errorf("app.taxicall.process [%s]: failed to handle taxi call progress command: %w", cmd.TaxiCallRequestId, err)
		}

		if err := t.repository.event.BatchCreate(ctx, i, events); err != nil {
			return fmt.Errorf("app.taxicall.process [%s]: failed to insert event: %w", cmd.TaxiCallRequestId, err)
		}

		return nil
	})
}

func (t taxicallApp) handleTaxiCallRequested(ctx context.Context, eventTime time.Time, receiveTime time.Time, taxiCallRequest entity.TaxiCallRequest) ([]entity.Event, error) {
	logger := utils.GetLogger(ctx)
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
				AdditionalPrice:   taxiCallRequest.RequestMinAdditionalPrice,
				TicketId:          utils.MustNewUUID(),
				CreateTime:        eventTime,
			}
		}

		if eventTime.Before(taxiCallTicket.CreateTime) {
			logger.Warn("event time is before than current ticket, may be late event arrived...",
				zap.String("type", "taxicall"),
				zap.String("method", "handleTaxiCallRequested"),
				zap.Any("currentRequest", taxiCallRequest),
				zap.Any("currentTicket", taxiCallTicket),
				zap.Any("eventTime", eventTime),
			)
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
			failedAnalytics := entity.NewAnalytics(receiveTime, analytics.UserTaxiCallRequestFailedPayload{
				Id:                        taxiCallRequest.Id,
				UserId:                    taxiCallRequest.UserId,
				FailedTime:                receiveTime,
				TaxiCallRequestCreateTime: taxiCallRequest.CreateTime,
			})
			if t.repository.analytics.Create(ctx, i, failedAnalytics); err != nil {
				return fmt.Errorf("app.taxicall.handleTaxiCallRequested: [%s] failed to create taxi call request failed analytics: %w", taxiCallRequest.Id, err)
			}
			events = append(events,
				command.NewTaxiCallProgressCommand(taxiCallRequest.Id, taxiCallRequest.CurrentState, receiveTime, receiveTime))
			return nil
		}

		// Get drivers
		driverTaxiCallContexts, err := t.repository.taxiCallRequest.
			GetDriverTaxiCallContextWithinRadius(ctx, i, taxiCallRequest.Departure, taxiCallRequest.Arrival, taxiCallTicket.GetRadius(), taxiCallRequest.TagIds,
				taxiCallTicket.TicketId, receiveTime)
		if err != nil {
			return fmt.Errorf("app.taxicall.handleTaxiCallRequested: [%s] error while get driver contexts within radius: %w", taxiCallRequest.Id, err)
		}
		driverTaxiCallContexts = selectTaxiCallContextsToDistribute(driverTaxiCallContexts)

		if len(driverTaxiCallContexts) == 0 && taxiCallTicket.AttemptLimit() {
			totalDistributedCount, err := t.repository.taxiCallRequest.GetDistributedCountByTicketId(ctx, i, taxiCallTicket.TicketId)
			if err != nil {
				return fmt.Errorf("app.taxicall.handleTaxiCallRequested [%s]: failed to get total distributed count: %w", taxiCallRequest.Id, err)
			}

			if totalDistributedCount == 0 {
				if err := taxiCallRequest.UpdateState(receiveTime, enum.TaxiCallState_DRIVER_NOT_AVAILABLE); err != nil {
					return fmt.Errorf("app.taxicall.handleTaxiCallRequested [%s]: failed to update state: %w", taxiCallRequest.Id, err)
				}
				if err := t.repository.taxiCallRequest.Update(ctx, i, taxiCallRequest); err != nil {
					return fmt.Errorf("app.taxicall.handleTaxiCallRequested: [%s] failed to update call request to failed state: %w", taxiCallRequest.Id, err)
				}

				driverNotAvailableAnalytics := entity.NewAnalytics(receiveTime, analytics.UserTaxiCallRequestDriverNotAvailablePayload{
					Id:                        taxiCallRequest.Id,
					UserId:                    taxiCallRequest.UserId,
					FailedTime:                receiveTime,
					TaxiCallRequestCreateTime: taxiCallRequest.CreateTime,
				})
				if t.repository.analytics.Create(ctx, i, driverNotAvailableAnalytics); err != nil {
					return fmt.Errorf("app.taxicall.handleTaxiCallRequested: [%s] failed to create taxi call request failed analytics: %w", taxiCallRequest.Id, err)
				}

				events = append(events,
					command.NewTaxiCallProgressCommand(taxiCallRequest.Id, taxiCallRequest.CurrentState, receiveTime, receiveTime))
				return nil
			}
		}

		if len(driverTaxiCallContexts) > 0 {
			// TODO (taekyeom) Tit-for-tat paramter 도 나중엔 configurable 하게 바꿔야 함
			// - Total entites: query로 부터 최대 몇 개의 entity를 가져 올 것인지
			// - Result entites: Total entites 중 몇 개의 entity에게 call ticket을 뿌릴 것인가
			// - Tit entites: top n 개의 entity를 몇 개 결정 할 것인지
			// - Tat entites (= Result entites - Tit entites) 몇 개의 entity를 total result 중 tit entity를 제외하고 랜덤하게 고를 것인지
			driverTaxiCallContexts = slices.Map(driverTaxiCallContexts, func(dctx entity.DriverTaxiCallContext) entity.DriverTaxiCallContext {
				dctx.LastReceivedRequestTicket = taxiCallTicket.TicketId
				dctx.LastReceiveTime = receiveTime
				dctx.RejectedLastRequestTicket = false
				return dctx
			})

			if err := t.repository.taxiCallRequest.BulkUpsertDriverTaxiCallContext(ctx, i, driverTaxiCallContexts); err != nil {
				return fmt.Errorf("app.taxicall.handleTaxiCallRequested: [%s] error while upsert driver contexts within radius: %w", taxiCallRequest.Id, err)
			}

			driverCmds := slices.Map(driverTaxiCallContexts, func(i entity.DriverTaxiCallContext) entity.Event {
				return command.NewPushDriverTaxiCallCommand(i.DriverId, taxiCallRequest, taxiCallTicket, i, receiveTime)
			})

			events = append(events, driverCmds...)

			taxiCallTicket.DistributedCount = len(driverTaxiCallContexts)

			ticketDistributionAnalytics := slices.Map(driverTaxiCallContexts, func(i entity.DriverTaxiCallContext) entity.Analytics {
				return entity.NewAnalytics(receiveTime, analytics.DriverTaxicallTicketDistributionPayload{
					DriverId:                  i.DriverId,
					RequestUserId:             taxiCallRequest.UserId,
					TaxiCallRequestId:         taxiCallRequest.Id,
					TaxiCallRequestTicketId:   taxiCallTicket.TicketId,
					TicketAttempt:             taxiCallTicket.Attempt,
					RequestBasePrice:          taxiCallRequest.RequestBasePrice,
					AdditionalPrice:           taxiCallTicket.AdditionalPrice,
					DriverLocation:            i.Location,
					TaxiCallRequestCreateTime: taxiCallRequest.CreateTime,
					DistanceToDeparture:       i.ToDepartureDistance,
					DistanceToArrival:         taxiCallRequest.ToDepartureRoute.Distance,
				})
			})
			if err := t.repository.analytics.BatchCreate(ctx, i, ticketDistributionAnalytics); err != nil {
				return fmt.Errorf("app.taxicall.handleTaxiCallRequested: [%s] error while create ticket distribution analytics events: %w", taxiCallRequest.Id, err)
			}
		}

		if err := t.repository.taxiCallRequest.CreateTicket(ctx, i, taxiCallTicket); err != nil {
			return fmt.Errorf("app.taxicall.handleTaxiCallRequested: [%s] error while create new ticket: %w", taxiCallRequest.Id, err)
		}

		userCmd := command.NewPushUserTaxiCallCommand(taxiCallRequest, taxiCallTicket, entity.DriverTaxiCallContext{}, receiveTime)

		taxiCallCmd := command.NewTaxiCallProgressCommand(taxiCallRequest.Id, taxiCallRequest.CurrentState,
			receiveTime, receiveTime.Add(time.Second*10))
		events = append(events, taxiCallCmd, userCmd)

		return nil
	})

	if err != nil {
		return []entity.Event{}, err
	}

	return events, nil
}

func selectTaxiCallContextsToDistribute(taxiCallContexts []entity.DriverTaxiCallContext) []entity.DriverTaxiCallContext {
	if len(taxiCallContexts) <= 5 {
		return taxiCallContexts
	} else {
		rand.Shuffle(len(taxiCallContexts)-2, func(i, j int) {
			taxiCallContexts[3+i], taxiCallContexts[3+j] = taxiCallContexts[3+j], taxiCallContexts[3+i]
		})
		return taxiCallContexts[:5]
	}
}

func (t taxicallApp) handleDriverToDeparture(ctx context.Context, eventTime time.Time, receiveTime time.Time, taxiCallRequest entity.TaxiCallRequest) ([]entity.Event, error) {
	events := []entity.Event{}
	err := t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		// TODO(taekyeom) Send location push message to user
		if eventTime.Before(taxiCallRequest.UpdateTime) {
			return nil
		}

		taxiCallTicket, err := t.repository.taxiCallRequest.GetLatestTicketByRequestId(ctx, i, taxiCallRequest.Id)
		if err != nil && !errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("app.taxicall.handleDriverToDeparture [%s]: error while get call request ticket: %w", taxiCallRequest.Id, err)
		}

		driverTaxiCallContext, err := t.repository.taxiCallRequest.GetDriverTaxiCallContext(ctx, i, taxiCallRequest.DriverId.String)
		if err != nil && !errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("app.taxicall.handleDriverToDeparture [%s]: error while get call request ticket: %w", taxiCallRequest.Id, err)
		}

		// TODO(taekyeom) 티켓 수신한 다른 기사분들을 다시 수신 가능한 상태로 만들어야 함
		ticketReceivedDriverContexts, err := t.repository.taxiCallRequest.GetDriverTaxiCallContextByTicketId(ctx, i, taxiCallTicket.TicketId)
		if err != nil {
			return fmt.Errorf("app.taxicall.handleDriverToDeparture [%s]: error while get taxi call contexts by tiket id: %w", taxiCallRequest.Id, err)
		}
		ticketReceivedDriverContexts = slices.Filter(ticketReceivedDriverContexts, func(i entity.DriverTaxiCallContext) bool {
			return i.DriverId != driverTaxiCallContext.DriverId
		})
		if len(ticketReceivedDriverContexts) > 0 {
			ticketReceivedDriverContexts = slices.Map(ticketReceivedDriverContexts, func(i entity.DriverTaxiCallContext) entity.DriverTaxiCallContext {
				// 다시 수신 가능한 상태를 rejection으로 처리
				i.RejectedLastRequestTicket = true
				return i
			})
			if err := t.repository.taxiCallRequest.BulkUpsertDriverTaxiCallContext(ctx, i, ticketReceivedDriverContexts); err != nil {
				return fmt.Errorf("app.taxicall.handleDriverToDeparture [%s]: error while bulk update driver contexts: %w", taxiCallRequest.Id, err)
			}
		}

		events = append(events, command.NewPushUserTaxiCallCommand(
			taxiCallRequest,
			taxiCallTicket,
			driverTaxiCallContext,
			receiveTime,
		))

		return nil
	})

	if err != nil {
		return []entity.Event{}, err
	}

	return events, nil
}

func (t taxicallApp) handleDriverToArrival(ctx context.Context, eventTime time.Time, receiveTime time.Time, taxiCallRequest entity.TaxiCallRequest) ([]entity.Event, error) {
	// TODO(taekyeom) Send location push message to user
	var events []entity.Event
	err := t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		events = append(events, command.NewPushUserTaxiCallCommand(taxiCallRequest, entity.TaxiCallTicket{}, entity.DriverTaxiCallContext{}, receiveTime))
		return nil
	})
	if err != nil {
		return []entity.Event{}, err
	}
	return events, nil
}

func (t taxicallApp) handleDone(ctx context.Context, eventTime time.Time, receiveTime time.Time, taxiCallRequest entity.TaxiCallRequest) ([]entity.Event, error) {
	logger := utils.GetLogger(ctx)
	if eventTime.Before(taxiCallRequest.UpdateTime) {
		logger.Warn("duplicated command...",
			zap.String("type", "taxicall"),
			zap.String("method", "handleDone"),
		)
		return []entity.Event{}, nil
	}

	events := []entity.Event{}
	err := t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		if err := t.repository.taxiCallRequest.DeleteTicketByRequestId(ctx, i, taxiCallRequest.Id); err != nil {
			return fmt.Errorf("app.taxicall.handleDone [%s]: failed to delete ticket: %w", taxiCallRequest.Id, err)
		}
		events = append(events, command.NewPushUserTaxiCallCommand(
			taxiCallRequest,
			entity.TaxiCallTicket{},
			entity.DriverTaxiCallContext{},
			receiveTime,
		))
		if taxiCallRequest.UserAdditionalPrice() > 0 {
			events = append(events, command.NewUserPaymentTransactionRequestCommand(
				taxiCallRequest.UserId,
				taxiCallRequest.PaymentSummary.PaymentId,
				taxiCallRequest.Id,
				"타코 이용 요금", // TODO (taekyeom) order name generation?
				taxiCallRequest.DriverId.String,
				taxiCallRequest.UserAdditionalPrice(),
				taxiCallRequest.UserUsedPoint,
				taxiCallRequest.DriverSettlementAdditonalPrice(),
				false,
			))
		}
		return nil
	})

	if err != nil {
		return []entity.Event{}, err
	}

	return events, nil
}

func (t taxicallApp) handleUserCancelled(ctx context.Context, eventTime time.Time, receiveTime time.Time, taxiCallRequest entity.TaxiCallRequest) ([]entity.Event, error) {
	logger := utils.GetLogger(ctx)
	if eventTime.Before(taxiCallRequest.UpdateTime) {
		logger.Warn("duplicated command...",
			zap.String("type", "taxicall"),
			zap.String("method", "handleUserCancelled"),
		)
		return []entity.Event{}, nil
	}

	events := []entity.Event{}
	err := t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		if err := t.repository.taxiCallRequest.DeleteTicketByRequestId(ctx, i, taxiCallRequest.Id); err != nil {
			return fmt.Errorf("app.taxicall.handleUserCancelled [%s]: failed to delete ticket: %w", taxiCallRequest.Id, err)
		}

		if taxiCallRequest.DriverId.Valid {
			events = append(events, command.NewPushDriverTaxiCallCommand(
				taxiCallRequest.DriverId.String,
				taxiCallRequest,
				entity.TaxiCallTicket{},
				entity.DriverTaxiCallContext{},
				receiveTime,
			))

			driverTaxiCallContext, err := t.repository.taxiCallRequest.GetDriverTaxiCallContext(ctx, i, taxiCallRequest.DriverId.String)
			if err != nil {
				return fmt.Errorf("app.taxicall.handleUserCancelled [%s]: error while get taxi call context: %w", taxiCallRequest.Id, err)
			}

			driverTaxiCallContext.CanReceive = true
			driverTaxiCallContext.RejectedLastRequestTicket = true
			if err := t.repository.taxiCallRequest.UpsertDriverTaxiCallContext(ctx, i, driverTaxiCallContext); err != nil {
				return fmt.Errorf("app.taxicall.handleUserCancelled [%s]: error while upsert taxi call context: %w", taxiCallRequest.Id, err)
			}
		}

		return nil
	})

	if err != nil {
		return []entity.Event{}, err
	}

	return events, nil
}

func (t taxicallApp) handleDriverCancelled(ctx context.Context, eventTime time.Time, receiveTime time.Time, taxiCallRequest entity.TaxiCallRequest) ([]entity.Event, error) {
	logger := utils.GetLogger(ctx)
	if eventTime.Before(taxiCallRequest.UpdateTime) {
		logger.Warn("duplicated command...",
			zap.String("type", "taxicall"),
			zap.String("method", "handleDriverCancelled"),
		)
		return []entity.Event{}, nil
	}

	events := []entity.Event{}
	err := t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		if err := t.repository.taxiCallRequest.DeleteTicketByRequestId(ctx, i, taxiCallRequest.Id); err != nil {
			return fmt.Errorf("app.taxicall.handleDriverCancelled [%s]: failed to delete ticket: %w", taxiCallRequest.Id, err)
		}

		driverTaxiCallContext, err := t.repository.taxiCallRequest.GetDriverTaxiCallContext(ctx, i, taxiCallRequest.DriverId.String)
		if err != nil {
			return fmt.Errorf("app.taxicall.handleDriverCancelled [%s]: error while get taxi call context: %w", taxiCallRequest.Id, err)
		}
		driverTaxiCallContext.CanReceive = true
		driverTaxiCallContext.RejectedLastRequestTicket = true
		if err := t.repository.taxiCallRequest.UpsertDriverTaxiCallContext(ctx, i, driverTaxiCallContext); err != nil {
			return fmt.Errorf("app.taxicall.handleDriverCancelled [%s]: error while upsert taxi call context: %w", taxiCallRequest.Id, err)
		}

		events = append(events, command.NewPushUserTaxiCallCommand(
			taxiCallRequest,
			entity.TaxiCallTicket{},
			entity.DriverTaxiCallContext{},
			receiveTime,
		))

		return nil
	})

	if err != nil {
		return []entity.Event{}, err
	}

	return events, nil
}

func (t taxicallApp) handleFailed(ctx context.Context, eventTime time.Time, receiveTime time.Time, taxiCallRequest entity.TaxiCallRequest) ([]entity.Event, error) {
	logger := utils.GetLogger(ctx)
	if eventTime.Before(taxiCallRequest.UpdateTime) {
		logger.Warn("duplicated command...",
			zap.String("type", "taxicall"),
			zap.String("method", "handleFailed"),
		)
		return []entity.Event{}, nil
	}

	events := []entity.Event{}
	err := t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		if err := t.repository.taxiCallRequest.DeleteTicketByRequestId(ctx, i, taxiCallRequest.Id); err != nil {
			return fmt.Errorf("app.taxicall.handleFailed [%s]: failed to delete ticket: %w", taxiCallRequest.Id, err)
		}
		events = append(events, command.NewPushUserTaxiCallCommand(
			taxiCallRequest,
			entity.TaxiCallTicket{},
			entity.DriverTaxiCallContext{},
			receiveTime,
		))
		return nil
	})

	if err != nil {
		return []entity.Event{}, err
	}

	return events, nil
}

func (t taxicallApp) handleDriverNotAvailable(ctx context.Context, eventTime time.Time, receiveTime time.Time, taxiCallRequest entity.TaxiCallRequest) ([]entity.Event, error) {
	logger := utils.GetLogger(ctx)
	if eventTime.Before(taxiCallRequest.UpdateTime) {
		logger.Warn("duplicated command...",
			zap.String("type", "taxicall"),
			zap.String("method", "handleDriverNotAvailable"),
		)
		return []entity.Event{}, nil
	}

	events := []entity.Event{}
	err := t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		if err := t.repository.taxiCallRequest.DeleteTicketByRequestId(ctx, i, taxiCallRequest.Id); err != nil {
			return fmt.Errorf("app.taxicall.handleDriverNotAvailable [%s]: failed to delete ticket: %w", taxiCallRequest.Id, err)
		}
		events = append(events, command.NewPushUserTaxiCallCommand(
			taxiCallRequest,
			entity.TaxiCallTicket{},
			entity.DriverTaxiCallContext{},
			receiveTime,
		))
		return nil
	})

	if err != nil {
		return []entity.Event{}, err
	}

	return events, nil
}
