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

func (u userApp) ListTaxiCallRequest(ctx context.Context, req request.ListUserTaxiCallRequest) ([]entity.TaxiCallRequest, string, error) {
	return u.service.taxiCall.ListUserTaxiCallRequest(ctx, req)
}

func (u userApp) GetLatestTaxiCallRequest(ctx context.Context, userId string) (entity.TaxiCallRequest, error) {
	return u.service.taxiCall.LatestUserTaxiCallRequest(ctx, userId)
}

func (u userApp) CreateTaxiCallRequest(ctx context.Context, req request.CreateTaxiCallRequest) (entity.TaxiCallRequest, error) {
	userId := utils.GetUserId(ctx)

	var userPayment entity.UserPayment
	err := u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		// check payment
		payment, err := u.repository.payment.GetUserPayment(ctx, i, req.PaymentId)
		if err != nil {
			return fmt.Errorf("app.user.CreateTaxiCallRequest: error while get user payment:\n%w", err)
		}

		if payment.UserId != userId {
			return fmt.Errorf("app.User.CreateTaxiCallRequest: unaurhorized payment:%w", value.ErrUnAuthorized)
		}

		userPayment = payment

		return nil
	})

	if err != nil {
		return entity.TaxiCallRequest{}, err
	}

	return u.service.taxiCall.CreateTaxiCallRequest(ctx, userId, userPayment, req)
}

func (u userApp) CancelTaxiCallRequest(ctx context.Context, taxiCallId string) error {
	userId := utils.GetUserId(ctx)
	return u.service.taxiCall.CancelTaxiCallRequest(ctx, userId, taxiCallId)
}
