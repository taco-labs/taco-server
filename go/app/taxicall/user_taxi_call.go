package taxicall

import (
	"context"
	"errors"
	"fmt"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/event/command"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/domain/value/analytics"
	"github.com/taco-labs/taco/go/domain/value/enum"
	"github.com/taco-labs/taco/go/utils"
	"github.com/taco-labs/taco/go/utils/slices"
	"github.com/uptrace/bun"
	"golang.org/x/sync/errgroup"
)

func (t taxicallApp) ListTags(ctx context.Context) ([]value.Tag, error) {
	var tags []value.Tag

	for id, tag := range value.TagMap {
		tags = append(tags, value.Tag{Id: id, Tag: tag})
	}

	return tags, nil
}

// TODO (taekyeom) userId / driverId 구분 없이 principal로 통합해야 할듯
func (t taxicallApp) ListUserTaxiCallRequest(ctx context.Context, req request.ListUserTaxiCallRequest) ([]entity.TaxiCallRequest, string, error) {
	var taxiCallRequests []entity.TaxiCallRequest
	var err error
	var pageToken string

	err = t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		taxiCallRequests, pageToken, err = t.repository.taxiCallRequest.ListByUserId(ctx, i, req.UserId, req.PageToken, req.Count)
		if err != nil {
			return fmt.Errorf("app.taxiCall.ListUserTaxiCallRequest: error while get taxi call requests:\n%w", err)
		}

		err := slices.ForeachErrRef(taxiCallRequests, func(i *entity.TaxiCallRequest) error {
			tags, err := slices.MapErr(i.TagIds, value.GetTagById)
			if err != nil {
				return err
			}
			i.Tags = tags
			return nil
		})
		if err != nil {
			return fmt.Errorf("app.taxiCall.ListUserTaxiCallRequest: error while get tags: %w", err)
		}

		return nil
	})

	if err != nil {
		return []entity.TaxiCallRequest{}, "", err
	}

	return taxiCallRequests, pageToken, nil
}

func (t taxicallApp) LatestUserTaxiCallRequest(ctx context.Context, userId string) (entity.UserLatestTaxiCallRequest, error) {
	var latestTaxiCallRequest entity.UserLatestTaxiCallRequest

	err := t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		taxiCallRequest, err := t.repository.taxiCallRequest.GetLatestByUserId(ctx, i, userId)
		if err != nil {
			return fmt.Errorf("app.taxiCall.LatestUserTaxiCallRequest: error while get latest taxi call:\n%w", err)
		}

		tags, err := slices.MapErr(taxiCallRequest.TagIds, value.GetTagById)
		if err != nil {
			return fmt.Errorf("app.taxiCall.LatestUserTaxiCallRequest: error while get tags: %w", err)
		}
		taxiCallRequest.Tags = tags

		var driver entity.Driver
		if taxiCallRequest.DriverId.Valid {
			driver, err = t.service.driverGetter.GetDriver(ctx, taxiCallRequest.DriverId.String)
			if err != nil {
				return fmt.Errorf("app.taxiCall.LatestUserTaxiCallRequest: error while get driver: %w", err)
			}
		} else {
			driver = entity.Driver{}
		}

		latestTaxiCallRequest = entity.UserLatestTaxiCallRequest{
			TaxiCallRequest: taxiCallRequest,
			DriverPhone:     driver.Phone,
			DriverCarNumber: driver.CarProfile.CarNumber,
		}

		return nil
	})

	if err != nil {
		return entity.UserLatestTaxiCallRequest{}, err
	}

	return latestTaxiCallRequest, nil
}

func (t taxicallApp) CreateTaxiCallRequest(ctx context.Context, user entity.User, req request.CreateTaxiCallRequest) (entity.TaxiCallRequest, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	tags, err := slices.MapErr(req.TagIds, value.GetTagById)
	if err != nil {
		return entity.TaxiCallRequest{}, fmt.Errorf("app.taxicall.CreateTaxiCallRequest: error while get tags: %w", err)
	}

	// First, check latest taxiCallRequest is active
	err = t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		// check latest call
		latestTaxiCallRequest, err := t.repository.taxiCallRequest.GetLatestByUserId(ctx, i, user.Id)
		if err != nil && !errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("app.taxiCall.CreateTaxiCallRequest: error while get latest taxi call:\n%w", err)
		}

		isNotFound := errors.Is(err, value.ErrNotFound)
		isActive := latestTaxiCallRequest.CurrentState.Active()

		if !isNotFound && isActive {
			err = fmt.Errorf("app.taxiCall.CreateTaxiCallRequest: already active taxi call request exists:\n%w", value.ErrAlreadyExists)
			return err
		}

		return nil
	})

	if err != nil {
		return entity.TaxiCallRequest{}, err
	}

	// get location of arrival, departure
	group, _ := errgroup.WithContext(ctx)
	var departure, arrival value.Address
	group.Go(func() error {
		addr, err := t.service.mapService.GetAddress(ctx, req.Departure)
		if err != nil {
			return fmt.Errorf("%w: error from get road address of departure", err)
		}
		departure = addr
		return nil
	})

	group.Go(func() error {
		addr, err := t.service.mapService.GetAddress(ctx, req.Arrival)
		if err != nil {
			return fmt.Errorf("%w: error from get road address of arrival", err)
		}
		arrival = addr
		return nil
	})

	if err := group.Wait(); err != nil {
		return entity.TaxiCallRequest{}, fmt.Errorf("app.taxiCall.CreateTaxiCallRequest: error while get %w", err)
	}

	isMockUser := user.MockAccount()

	departureAvailableRegion, err := t.service.userServiceRegionChecker.CheckAvailableServiceRegion(ctx, departure.ServiceRegion)
	if err != nil {
		return entity.TaxiCallRequest{}, fmt.Errorf("app.taxiCall.CreateTaxiCallRequest: error while check departure service region:\n%w", err)
	}
	arrivalAvailableRegion, err := t.service.userServiceRegionChecker.CheckAvailableServiceRegion(ctx, arrival.ServiceRegion)
	if err != nil {
		return entity.TaxiCallRequest{}, fmt.Errorf("app.taxiCall.CreateTaxiCallRequest: error while check arrival service region:\n%w", err)
	}

	// TODO(taekyeom) To be paramterized
	if !isMockUser && !(departureAvailableRegion || arrivalAvailableRegion) {
		return entity.TaxiCallRequest{}, fmt.Errorf("app.taxiCall.CreateTaxiCallRequest: not supported region: %w", value.ErrUnsupportedServiceRegion)
	}

	route, err := t.service.mapService.GetRoute(ctx, req.Departure, req.Arrival)
	if err != nil {
		return entity.TaxiCallRequest{}, fmt.Errorf("app.taxiCall.CreateTaxiCallRequest: error while get route:\n%w", err)
	}

	if req.Dryrun {
		// TODO (taekyeom) additional price 계산을 위한 별도 모듈로 빼야 함
		maxAdditionalPrice := 10000
		taxiCallRequest := entity.TaxiCallRequest{
			Dryrun: req.Dryrun,
			UserId: user.Id,
			ToArrivalRoute: entity.TaxiCallToArrivalRoute{
				Route: route,
			},
			Departure: value.Location{
				Point:   req.Departure,
				Address: departure,
			},
			Arrival: value.Location{
				Point:   req.Arrival,
				Address: arrival,
			},
			TagIds:                    req.TagIds,
			Tags:                      tags,
			UserTag:                   req.UserTag,
			RequestBasePrice:          route.Price,
			RequestMinAdditionalPrice: 0,                  // TODO(taekyeom) To be paramterized
			RequestMaxAdditionalPrice: maxAdditionalPrice, // TODO(taekyeom) To be paramterized
			CurrentState:              enum.TaxiCallState_DRYRUN,
			CreateTime:                requestTime,
			UpdateTime:                requestTime,
		}

		return taxiCallRequest, nil
	}

	var taxiCallRequest entity.TaxiCallRequest

	err = t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		// create taxi call request
		userPayment, err := t.service.payment.GetUserPayment(ctx, user.Id, req.PaymentId)
		if err != nil {
			return fmt.Errorf("app.taxiCall.CreateTaxiCallRequest: error while get user payment: %w", err)
		}

		taxiCallRequestId := utils.MustNewUUID()

		toArrivalRoute := entity.TaxiCallToArrivalRoute{
			TaxiCallRequestId: taxiCallRequestId,
			Route:             route,
		}

		taxiCallRequest = entity.TaxiCallRequest{
			Id:             taxiCallRequestId,
			Dryrun:         req.Dryrun,
			ToArrivalRoute: toArrivalRoute,
			UserId:         user.Id,
			Departure: value.Location{
				Point:   req.Departure,
				Address: departure,
			},
			Arrival: value.Location{
				Point:   req.Arrival,
				Address: arrival,
			},
			TagIds:                    req.TagIds,
			Tags:                      tags,
			UserTag:                   req.UserTag,
			PaymentSummary:            userPayment.ToSummary(),
			RequestBasePrice:          route.Price,
			RequestMinAdditionalPrice: req.MinAdditionalPrice,
			RequestMaxAdditionalPrice: req.MaxAdditionalPrice,
			CurrentState:              enum.TaxiCallState_Requested,
			CreateTime:                requestTime,
			UpdateTime:                requestTime,
		}

		if err := entity.CheckPaymentForTaxiCallRequest(&taxiCallRequest, userPayment); err != nil {
			return fmt.Errorf("app.taxiCall.CreateTaxiCallRequest: error while check payment: %w", err)
		}

		userPayment.LastUseTime = requestTime
		// TODO (taekyeom) centralize promotion payment operations...
		if userPayment.PaymentType == enum.PaymentType_SignupPromition {
			userPayment.Invalid = true
			userPayment.InvalidErrorMessage = "이미 사용한 프로모션입니다"
		}
		if err := t.service.payment.UpdateUserPayment(ctx, userPayment); err != nil {
			return fmt.Errorf("app.taxiCall.CreateTaxiCallRequest: error while update user payment: %w", err)
		}
		if err = t.repository.taxiCallRequest.Create(ctx, i, taxiCallRequest); err != nil {
			return fmt.Errorf("app.taxiCall.CreateTaxiCallRequest: error while create taxi call request:%w", err)
		}
		if err := t.repository.taxiCallRequest.CreateToArrivalRoute(ctx, i, toArrivalRoute); err != nil {
			return fmt.Errorf("app.taxiCall.CreateTaxiCallRequest: error while create taxi call to arrival route:%w", err)
		}

		processMessage := command.NewTaxiCallProgressCommand(
			taxiCallRequest.Id,
			taxiCallRequest.CurrentState,
			taxiCallRequest.UpdateTime,
			taxiCallRequest.UpdateTime,
		)

		if err := t.repository.event.BatchCreate(ctx, i, []entity.Event{processMessage}); err != nil {
			return fmt.Errorf("app.taxiCall.CreateTaxiCallRequest: error while create taxi call process event: %w", err)
		}

		requestAnalytics := entity.NewAnalytics(requestTime, analytics.UserTaxiCallRequestPayload{
			UserId:                    taxiCallRequest.UserId,
			Id:                        taxiCallRequest.Id,
			Departure:                 taxiCallRequest.Departure,
			Arrival:                   taxiCallRequest.Arrival,
			ToArrivalETA:              taxiCallRequest.ToArrivalRoute.Route.ETA,
			ToArrivalDistance:         taxiCallRequest.ToArrivalRoute.Route.Distance,
			TagIds:                    taxiCallRequest.TagIds,
			Tags:                      taxiCallRequest.Tags,
			UserTag:                   taxiCallRequest.UserTag,
			PaymentSummary:            taxiCallRequest.PaymentSummary,
			RequestBasePrice:          taxiCallRequest.RequestBasePrice,
			RequestMinAdditionalPrice: taxiCallRequest.RequestMinAdditionalPrice,
			RequestMaxAdditionalPrice: taxiCallRequest.RequestMaxAdditionalPrice,
		})
		if err := t.repository.analytics.Create(ctx, i, requestAnalytics); err != nil {
			return fmt.Errorf("app.taxiCall.CreateTaxiCallRequest: error while create analytics event: %w", err)
		}

		return nil
	})

	if err != nil {
		return entity.TaxiCallRequest{}, err
	}

	return taxiCallRequest, nil
}

func (t taxicallApp) UserCancelTaxiCallRequest(ctx context.Context, userId string, req request.UserCancelTaxiCallRequest) error {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	return t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		var events []entity.Event
		taxiCallRequest, err := t.repository.taxiCallRequest.GetById(ctx, i, req.TaxiCallRequestId)
		if err != nil {
			return fmt.Errorf("app.taxiCall.CancelTaxiCall: error while get taxi call:%w", err)
		}

		if taxiCallRequest.UserId != userId {
			return fmt.Errorf("app.taxiCall.CancelTaxiCall: Invalid request:%w", value.ErrUnAuthorized)
		}

		err = taxiCallRequest.UpdateState(requestTime, enum.TaxiCallState_USER_CANCELLED)
		if err != nil && !errors.Is(err, value.ErrConfirmationNeededStateTransition) {
			return err
		}
		if errors.Is(err, value.ErrConfirmationNeededStateTransition) && !req.ConfirmCancel {
			return err
		}
		if errors.Is(err, value.ErrConfirmationNeededStateTransition) && req.ConfirmCancel {
			taxiCallRequest.ForceUpdateState(requestTime, enum.TaxiCallState_USER_CANCELLED)
			taxiCallRequest.CancelPenaltyPrice = taxiCallRequest.UserCancelPenaltyPrice(requestTime)
		}

		if err = t.repository.taxiCallRequest.Update(ctx, i, taxiCallRequest); err != nil {
			return fmt.Errorf("app.taxiCall.CancelTaxiCall: error while update taxi call:%w", err)
		}

		if err := t.service.payment.AddUserPaymentPoint(ctx, userId, taxiCallRequest.UserUsedPoint); err != nil {
			return fmt.Errorf("app.taxiCall.CancelTaxiCall: error while add user used payment point: %w", err)
		}

		events = append(events, command.NewTaxiCallProgressCommand(
			taxiCallRequest.Id,
			taxiCallRequest.CurrentState,
			taxiCallRequest.UpdateTime,
			taxiCallRequest.UpdateTime,
		))

		if taxiCallRequest.CancelPenaltyPrice == 0 && taxiCallRequest.PaymentSummary.PaymentType == enum.PaymentType_SignupPromition {
			userPayment, err := t.service.payment.GetUserPayment(ctx, taxiCallRequest.UserId, taxiCallRequest.PaymentSummary.PaymentId)
			if err != nil {
				return fmt.Errorf("app.taxiCall.DriverCancelTaxiCall: error while get user payment: %w", err)
			}
			userPayment.Invalid = false
			userPayment.InvalidErrorMessage = ""
			if err := t.service.payment.UpdateUserPayment(ctx, userPayment); err != nil {
				return fmt.Errorf("app.taxiCall.DriverCancelTaxiCall: error while update user payment: %w", err)
			}
		}

		if taxiCallRequest.CancelPenaltyPrice > 0 {
			events = append(events, command.NewUserPaymentTransactionRequestCommand(
				taxiCallRequest.UserId,
				taxiCallRequest.PaymentSummary.PaymentId,
				taxiCallRequest.Id,
				"타코 택시 취소 수수료",
				taxiCallRequest.DriverId.String,
				taxiCallRequest.CancelPenaltyPrice,
				0,
				taxiCallRequest.DriverSettlementCancelPenaltyPrice(),
				false,
			))
		}

		if err := t.repository.event.BatchCreate(ctx, i, events); err != nil {
			return fmt.Errorf("app.taxiCall.CreateTaxiCallRequest: error while create taxi call process event: %w", err)
		}

		cancelAnalytics := entity.NewAnalytics(requestTime, analytics.UserCancelTaxiCallRequestPayload{
			UserId:                    taxiCallRequest.UserId,
			Id:                        taxiCallRequest.Id,
			CancelPenalty:             taxiCallRequest.UserCancelPenaltyPrice(requestTime),
			TaxiCallRequestCreateTime: taxiCallRequest.CreateTime,
		})
		if err := t.repository.analytics.Create(ctx, i, cancelAnalytics); err != nil {
			return fmt.Errorf("app.taxiCall.CancelTaxiCall: error while create analytics event: %w", err)
		}

		return nil
	})
}
