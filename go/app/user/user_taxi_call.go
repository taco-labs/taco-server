package user

import (
	"fmt"

	"context"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/utils"
	"github.com/uptrace/bun"
)

func (u userApp) ListTags(ctx context.Context) ([]value.Tag, error) {
	return u.service.taxiCall.ListTags(ctx)
}

func (u userApp) ListTaxiCallRequest(ctx context.Context, req request.ListUserTaxiCallRequest) ([]entity.TaxiCallRequest, string, error) {
	return u.service.taxiCall.ListUserTaxiCallRequest(ctx, req)
}

func (u userApp) GetLatestTaxiCallRequest(ctx context.Context, userId string) (entity.UserLatestTaxiCallRequest, error) {
	return u.service.taxiCall.LatestUserTaxiCallRequest(ctx, userId)
}

func (u userApp) CreateTaxiCallRequest(ctx context.Context, req request.CreateTaxiCallRequest) (entity.TaxiCallRequest, error) {
	if err := req.Validate(); err != nil {
		return entity.TaxiCallRequest{}, fmt.Errorf("app.user.CreateTaxiCallRequest: error while validation: %w", err)
	}

	if !u.config.taxiCallEnabled {
		return entity.TaxiCallRequest{}, fmt.Errorf("app.user.CreateTaxiCallRequest: temprarily shutdown: %w", value.ErrTemporarilyUnsupported)
	}

	userId := utils.GetUserId(ctx)
	var user entity.User

	err := u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		u, err := u.repository.user.FindById(ctx, i, userId)
		if err != nil {
			return fmt.Errorf("app.user.CreateTaxiCallRequest: error while get user: %w", err)
		}
		user = u
		return nil
	})

	if err != nil {
		return entity.TaxiCallRequest{}, err
	}

	taxiCallRequest, err := u.service.taxiCall.CreateTaxiCallRequest(ctx, user, req)
	if err != nil {
		return entity.TaxiCallRequest{}, fmt.Errorf("app.user.CreateTaxiCallRequest: error while create taxi call request:%w", err)
	}

	return taxiCallRequest, nil
}

func (u userApp) CancelTaxiCallRequest(ctx context.Context, req request.UserCancelTaxiCallRequest) error {
	userId := utils.GetUserId(ctx)
	err := u.service.taxiCall.UserCancelTaxiCallRequest(ctx, userId, req)
	if err != nil {
		return fmt.Errorf("app.user.CancelTaxiCallRequest: error while cancel taxi call request:%w", err)
	}

	return nil
}

func (u userApp) GetUserLatestTaxiCallTicket(ctx context.Context, taxiCallRequestId string) (entity.TaxiCallTicket, error) {
	userId := utils.GetUserId(ctx)
	return u.service.taxiCall.GetUserLatestTaxiCallTicket(ctx, userId, taxiCallRequestId)
}
