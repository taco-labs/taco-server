package user

import (
	"errors"
	"fmt"

	"context"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/domain/value/enum"
	"github.com/taco-labs/taco/go/utils"
	"github.com/uptrace/bun"
	"golang.org/x/sync/errgroup"
)

func (u userApp) ListTaxiCallRequest(ctx context.Context, req request.ListUserTaxiCallRequest) ([]entity.TaxiCallRequest, string, error) {
	var taxiCallRequests []entity.TaxiCallRequest
	var err error
	var pageToken string

	err = u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		taxiCallRequests, pageToken, err = u.repository.taxiCallRequest.ListByUserId(ctx, i, req.UserId, req.PageToken, req.Count)
		if err != nil {
			return fmt.Errorf("app.user.ListTaxiCallRequest: error while get taxi call requests:\n%w", err)
		}
		return nil
	})

	if err != nil {
		return []entity.TaxiCallRequest{}, "", nil
	}

	return taxiCallRequests, pageToken, nil
}

func (u userApp) GetLatestTaxiCallRequest(ctx context.Context, userId string) (entity.TaxiCallRequest, error) {
	var latestTaxiCallRequest entity.TaxiCallRequest
	var err error

	err = u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		latestTaxiCallRequest, err = u.repository.taxiCallRequest.GetLatestByUserId(ctx, i, userId)
		if err != nil {
			return fmt.Errorf("app.user.GetLatestTaxiCall: error while get latest taxi call:\n%w", err)
		}
		return nil
	})

	if err != nil {
		return entity.TaxiCallRequest{}, err
	}

	return latestTaxiCallRequest, nil
}

func (u userApp) CreateTaxiCallRequest(ctx context.Context, req request.CreateTaxiCallRequest) (entity.TaxiCallRequest, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)
	userId := utils.GetUserId(ctx)

	// get location of arrival, departure
	group, _ := errgroup.WithContext(ctx)
	var departure, arrival value.Address
	group.Go(func() error {
		roadAddress, err := u.service.location.GetAddress(ctx, req.Departure)
		if err != nil {
			return fmt.Errorf("%w: error from get road address of departure", err)
		}
		departure = roadAddress
		return nil
	})

	group.Go(func() error {
		roadAddress, err := u.service.location.GetAddress(ctx, req.Arrival)
		if err != nil {
			return fmt.Errorf("%w: error from get road address of arrival", err)
		}
		arrival = roadAddress
		return nil
	})

	if err := group.Wait(); err != nil {
		return entity.TaxiCallRequest{}, fmt.Errorf("%w: error from get address", err)
	}

	// TODO(taekyeom) To be paramterized
	if !(departure.AvailableRegion() && arrival.AvailableRegion()) {
		return entity.TaxiCallRequest{}, fmt.Errorf("%w: not supported region", value.ErrUnsupportedRegion)
	}

	route, err := u.service.route.GetRoute(ctx, req.Departure, req.Arrival)
	if err != nil {
		return entity.TaxiCallRequest{}, fmt.Errorf("app.user.CreateTaxiCallRequest: error while get route:\n%w", err)
	}

	var taxiCallRequest entity.TaxiCallRequest
	err = u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		// check latest call
		latestTaxiCallRequest, err := u.repository.taxiCallRequest.GetLatestByUserId(ctx, i, userId)
		if err != nil && !errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("app.user.CreateTaxiCallRequest: error while get latest taxi call:\n%w", err)
		}

		isNotFound := errors.Is(err, value.ErrNotFound)
		isActive := latestTaxiCallRequest.CurrentState.Active()

		if !isNotFound && isActive {
			err = fmt.Errorf("app.user.CreateTaxiCallRequest: already active taxi call request exists:\n%w", value.ErrAlreadyExists)
			return err
		}

		if req.Dryrun {
			taxiCallRequest = entity.TaxiCallRequest{
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

			return nil
		}

		// check payment
		userPayment, err := u.repository.payment.GetUserPayment(ctx, i, req.PaymentId)
		if err != nil {
			return fmt.Errorf("app.user.CreateTaxiCallRequest: error while get user payment:\n%w", err)
		}

		if userPayment.UserId != userId {
			return fmt.Errorf("app.User.CreateTaxiCallRequest: unaurhorized payment:%w", value.ErrUnAuthorized)
		}

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

		if err = u.repository.taxiCallRequest.Create(ctx, i, taxiCallRequest); err != nil {
			return fmt.Errorf("app.user.CreateTaxiCallRequest: error while create taxi call request:%w", err)
		}

		return nil
	})

	if err != nil {
		return entity.TaxiCallRequest{}, err
	}

	if !req.Dryrun {
		if err = u.actor.taxiCallRequest.Add(taxiCallRequest.Id); err != nil {
			return entity.TaxiCallRequest{}, fmt.Errorf("app.user.CreateTaxiCallRequest: error while adding actor:%w", err)
		}
	}

	return taxiCallRequest, nil
}

func (u userApp) CancelTaxiCallRequest(ctx context.Context, taxiCallId string) error {
	requestTime := utils.GetRequestTimeOrNow(ctx)
	userId := utils.GetUserId(ctx)

	return u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		taxiCall, err := u.repository.taxiCallRequest.GetById(ctx, i, taxiCallId)
		if err != nil {
			return fmt.Errorf("app.user.CancelTaxiCall: error while get taxi call:%w", err)
		}

		if taxiCall.UserId != userId {
			return fmt.Errorf("app.user.CancelTaxiCall: Invalid request:%w", value.ErrUnAuthorized)
		}

		if err = taxiCall.UpdateState(requestTime, enum.TaxiCallState_USER_CANCELLED); err != nil {
			return fmt.Errorf("app.user.CancelTaxiCall: error while cancel taxi call:%w", err)
		}

		if err = u.repository.taxiCallRequest.Update(ctx, i, taxiCall); err != nil {
			return fmt.Errorf("app.user.CancelTaxiCall: error while update taxi call:%w", err)
		}

		return nil
	})
}
