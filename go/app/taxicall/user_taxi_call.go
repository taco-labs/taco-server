package taxicall

import (
	"context"
	"errors"
	"fmt"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/event/command"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/domain/value/enum"
	"github.com/taco-labs/taco/go/utils"
	"github.com/uptrace/bun"
	"golang.org/x/sync/errgroup"
)

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
		return nil
	})

	if err != nil {
		return []entity.TaxiCallRequest{}, "", nil
	}

	return taxiCallRequests, pageToken, nil
}

func (t taxicallApp) LatestUserTaxiCallRequest(ctx context.Context, userId string) (entity.TaxiCallRequest, error) {
	var latestTaxiCallRequest entity.TaxiCallRequest
	var err error

	err = t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		latestTaxiCallRequest, err = t.repository.taxiCallRequest.GetLatestByUserId(ctx, i, userId)
		if err != nil {
			return fmt.Errorf("app.taxCall.GetLatestTaxiCall: error while get latest taxi call:\n%w", err)
		}
		return nil
	})

	if err != nil {
		return entity.TaxiCallRequest{}, err
	}

	return latestTaxiCallRequest, nil
}

func (t taxicallApp) CreateTaxiCallRequest(ctx context.Context, userId string, userPayment entity.UserPayment, req request.CreateTaxiCallRequest) (entity.TaxiCallRequest, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	// First, check latest taxiCallRequest is active
	err := t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		// check latest call
		latestTaxiCallRequest, err := t.repository.taxiCallRequest.GetLatestByUserId(ctx, i, userId)
		if err != nil && !errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("app.taxiCall.CreateTaxiCallRequest: error while get latest taxi call:\n%w", err)
		}

		isNotFound := errors.Is(err, value.ErrNotFound)
		isActive := latestTaxiCallRequest.CurrentState.Active()

		if !isNotFound && isActive {
			err = fmt.Errorf("app.taxCall.CreateTaxiCallRequest: already active taxi call request exists:\n%w", value.ErrAlreadyExists)
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
		roadAddress, err := t.service.location.GetAddress(ctx, req.Departure)
		if err != nil {
			return fmt.Errorf("%w: error from get road address of departure", err)
		}
		departure = roadAddress
		return nil
	})

	group.Go(func() error {
		roadAddress, err := t.service.location.GetAddress(ctx, req.Arrival)
		if err != nil {
			return fmt.Errorf("%w: error from get road address of arrival", err)
		}
		arrival = roadAddress
		return nil
	})

	if err := group.Wait(); err != nil {
		return entity.TaxiCallRequest{}, fmt.Errorf("app.taxiCall.CreateTaxiCallRequest: error while get %w", err)
	}

	// TODO(taekyeom) To be paramterized
	if !(departure.AvailableRegion() && arrival.AvailableRegion()) {
		return entity.TaxiCallRequest{}, fmt.Errorf("%w: not supported region", value.ErrUnsupportedRegion)
	}

	route, err := t.service.route.GetRoute(ctx, req.Departure, req.Arrival)
	if err != nil {
		return entity.TaxiCallRequest{}, fmt.Errorf("app.taxCall.CreateTaxiCallRequest: error while get route:\n%w", err)
	}

	if req.Dryrun {
		taxiCallRequest := entity.TaxiCallRequest{
			Dryrun: req.Dryrun,
			UserId: userId,
			Route:  route,
			Departure: value.Location{
				Point:   req.Departure,
				Address: departure,
			},
			Arrival: value.Location{
				Point:   req.Arrival,
				Address: arrival,
			},
			RequestBasePrice:          route.Price,
			RequestMinAdditionalPrice: 0,                           // TODO(taekyeom) To be paramterized
			RequestMaxAdditionalPrice: (route.Price / 1000) * 1000, // TODO(taekyeom) To be paramterized
			CurrentState:              enum.TaxiCallState_DRYRUN,
			CreateTime:                requestTime,
			UpdateTime:                requestTime,
		}

		return taxiCallRequest, nil
	}

	var taxiCallRequest entity.TaxiCallRequest

	err = t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		// create taxi call request
		taxiCallRequest = entity.TaxiCallRequest{
			Dryrun: req.Dryrun,
			Route:  route,
			Id:     utils.MustNewUUID(),
			UserId: userId,
			Departure: value.Location{
				Point:   req.Departure,
				Address: departure,
			},
			Arrival: value.Location{
				Point:   req.Arrival,
				Address: arrival,
			},
			PaymentSummary: value.PaymentSummary{
				PaymentId:  userPayment.Id,
				Company:    userPayment.CardCompany,
				CardNumber: userPayment.RedactedCardNumber,
			},
			RequestBasePrice:          route.Price,
			RequestMinAdditionalPrice: 0,           // TODO(taekyeom) To be paramterized
			RequestMaxAdditionalPrice: route.Price, // TODO(taekyeom) To be paramterized
			CurrentState:              enum.TaxiCallState_Requested,
			CreateTime:                requestTime,
			UpdateTime:                requestTime,
		}

		if err = t.repository.taxiCallRequest.Create(ctx, i, taxiCallRequest); err != nil {
			return fmt.Errorf("app.taxCall.CreateTaxiCallRequest: error while create taxi call request:%w", err)
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

		return nil
	})

	if err != nil {
		return entity.TaxiCallRequest{}, err
	}

	return taxiCallRequest, nil
}

func (t taxicallApp) CancelTaxiCallRequest(ctx context.Context, userId string, taxiCallId string) error {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	return t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		taxiCallRequest, err := t.repository.taxiCallRequest.GetById(ctx, i, taxiCallId)
		if err != nil {
			return fmt.Errorf("app.taxCall.CancelTaxiCall: error while get taxi call:%w", err)
		}

		if taxiCallRequest.UserId != userId {
			return fmt.Errorf("app.taxCall.CancelTaxiCall: Invalid request:%w", value.ErrUnAuthorized)
		}

		if err = taxiCallRequest.UpdateState(requestTime, enum.TaxiCallState_USER_CANCELLED); err != nil {
			return fmt.Errorf("app.taxCall.CancelTaxiCall: error while cancel taxi call:%w", err)
		}

		if err = t.repository.taxiCallRequest.Update(ctx, i, taxiCallRequest); err != nil {
			return fmt.Errorf("app.taxCall.CancelTaxiCall: error while update taxi call:%w", err)
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

		return nil
	})
}
