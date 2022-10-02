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
)

func (u userApp) ListTaxiCallRequest(ctx context.Context, userId string) ([]entity.TaxiCallRequest, error) {
	ctx, err := u.Start(ctx)
	if err != nil {
		return []entity.TaxiCallRequest{}, err
	}
	defer func() {
		err = u.Done(ctx, err)
	}()

	taxiCallRequests, err := u.repository.taxiCallRequest.ListByUserId(ctx, userId)
	if err != nil {
		return []entity.TaxiCallRequest{}, fmt.Errorf("app.user.ListTaxiCallRequest: error while get taxi call requests:\n%w", err)
	}

	return taxiCallRequests, nil
}

func (u userApp) GetLatestTaxiCallRequest(ctx context.Context, userId string) (entity.TaxiCallRequest, error) {
	ctx, err := u.Start(ctx)
	if err != nil {
		return entity.TaxiCallRequest{}, err
	}
	defer func() {
		err = u.Done(ctx, err)
	}()

	latestTaxiCallRequest, err := u.repository.taxiCallRequest.GetLatestByUserId(ctx, userId)
	if err != nil {
		return entity.TaxiCallRequest{}, fmt.Errorf("app.user.GetLatestTaxiCall: error while get latest taxi call:\n%w", err)
	}

	return latestTaxiCallRequest, nil
}

func (u userApp) CreateTaxiCallRequest(ctx context.Context, req request.CreateTaxiCallRequest) (entity.TaxiCallRequest, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)
	userId := utils.GetUserId(ctx)

	ctx, err := u.Start(ctx)
	if err != nil {
		return entity.TaxiCallRequest{}, err
	}
	defer func() {
		err = u.Done(ctx, err)
	}()

	// check latest call
	latestTaxiCallRequest, err := u.repository.taxiCallRequest.GetLatestByUserId(ctx, userId)
	if err != nil && !errors.Is(err, value.ErrNotFound) {
		return entity.TaxiCallRequest{}, fmt.Errorf("app.user.CreateTaxiCallRequest: error while get latest taxi call:\n%w", err)
	}

	isNotFound := errors.Is(err, value.ErrNotFound)
	isActive := latestTaxiCallRequest.CurrentState.Active()

	if !isNotFound && isActive {
		err = fmt.Errorf("app.user.CreateTaxiCallRequest: already active taxi call request exists:\n%w", value.ErrAlreadyExists)
		return entity.TaxiCallRequest{}, err
	}

	// check payment
	userPayment, err := u.repository.payment.GetUserPayment(ctx, req.PaymentId)
	if err != nil {
		return entity.TaxiCallRequest{}, fmt.Errorf("app.user.CreateTaxiCallRequest: error while get user payment:\n%w", err)
	}

	if userPayment.UserId != userId {
		err = fmt.Errorf("app.User.CreateTaxiCallRequest: unaurhorized payment:%w", value.ErrUnAuthorized)
		return entity.TaxiCallRequest{}, err
	}

	// Get route
	route, err := u.service.route.GetRoute(ctx, req.Departure, req.Arrival)
	if err != nil {
		return entity.TaxiCallRequest{}, fmt.Errorf("app.user.CreateTaxiCallRequest: error while get route:\n%w", err)
	}

	// create taxi call request
	initialState := enum.TaxiCallState_Requested
	taxiCallRequest := entity.TaxiCallRequest{
		Dryrun:    req.Dryrun,
		Id:        utils.MustNewUUID(),
		UserId:    userId,
		Departure: req.Departure,
		Arrival:   req.Arrival,
		PaymentSummary: value.PaymentSummary{
			PaymentId:  userPayment.Id,
			Company:    userPayment.CardCompany,
			CardNumber: userPayment.RedactedCardNumber,
		},
		RequestBasePrice:          route.Price,
		RequestMinAdditionalPrice: 0,           // TODO(taekyeom) To be paramterized
		RequestMaxAdditionalPrice: route.Price, // TODO(taekyeom) To be paramterized
		CurrentState:              initialState,
		CreateTime:                requestTime,
		UpdateTime:                requestTime,
	}

	if taxiCallRequest.Dryrun {
		return taxiCallRequest, nil
	}

	if err = u.repository.taxiCallRequest.Create(ctx, taxiCallRequest); err != nil {
		return entity.TaxiCallRequest{}, fmt.Errorf("app.user.CreateTaxiCallRequest: error while create taxi call request:%w", err)
	}

	return taxiCallRequest, nil
}

func (u userApp) CancelTaxiCallRequest(ctx context.Context, taxiCallId string) error {
	requestTime := utils.GetRequestTimeOrNow(ctx)
	userId := utils.GetUserId(ctx)

	ctx, err := u.Start(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = u.Done(ctx, err)
	}()

	taxiCall, err := u.repository.taxiCallRequest.GetById(ctx, taxiCallId)
	if err != nil {
		return fmt.Errorf("app.user.CancelTaxiCall: error while get taxi call:%w", err)
	}

	if taxiCall.UserId != userId {
		return fmt.Errorf("app.user.CancelTaxiCall: Invalid request:%w", value.ErrUnAuthorized)
	}

	if err = taxiCall.UpdateState(requestTime, enum.TaxiCallState_USER_CANCELLED); err != nil {
		return fmt.Errorf("app.user.CancelTaxiCall: error while cancel taxi call:%w", err)
	}

	if err = u.repository.taxiCallRequest.Update(ctx, taxiCall); err != nil {
		return fmt.Errorf("app.user.CancelTaxiCall: error while update taxi call:%w", err)
	}

	return nil
}
