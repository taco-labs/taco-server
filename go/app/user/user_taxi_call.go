package user

import (
	"fmt"

	"context"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/utils"
)

func (u userApp) ListTags(ctx context.Context) ([]value.Tag, error) {
	return u.service.taxiCall.ListTags(ctx)
}

func (u userApp) ListTaxiCallRequest(ctx context.Context, req request.ListUserTaxiCallRequest) ([]entity.TaxiCallRequest, string, error) {
	return u.service.taxiCall.ListUserTaxiCallRequest(ctx, req)
}

func (u userApp) GetLatestTaxiCallRequest(ctx context.Context, userId string) (entity.TaxiCallRequest, error) {
	return u.service.taxiCall.LatestUserTaxiCallRequest(ctx, userId)
}

func (u userApp) CreateTaxiCallRequest(ctx context.Context, req request.CreateTaxiCallRequest) (entity.TaxiCallRequest, error) {
	if err := req.Validate(); err != nil {
		return entity.TaxiCallRequest{}, fmt.Errorf("app.user.CreateTaxiCallRequest: error while validation: %w", err)
	}

	userId := utils.GetUserId(ctx)

	userPayment, err := u.service.userPayment.GetUserPayment(ctx, userId, req.PaymentId)
	if err != nil {
		return entity.TaxiCallRequest{}, fmt.Errorf("app.user.CreateTaxiCallRequest: error while get user payment:\n%w", err)
	}

	return u.service.taxiCall.CreateTaxiCallRequest(ctx, userId, userPayment, req)
}

func (u userApp) CancelTaxiCallRequest(ctx context.Context, taxiCallId string) error {
	userId := utils.GetUserId(ctx)
	return u.service.taxiCall.UserCancelTaxiCallRequest(ctx, userId, taxiCallId)
}
