package user

import (
	"context"
	"fmt"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/utils"
	"github.com/uptrace/bun"
)

func (u userApp) GetCardRegistrationRequestParam(ctx context.Context, userId string) (value.PaymentRegistrationRequestParam, error) {
	var user entity.User
	err := u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		u, err := u.repository.user.FindById(ctx, i, userId)
		if err != nil {
			return fmt.Errorf("app.user.GetCardRegistrationRequestParam: error while get user: %w", err)
		}

		user = u
		return nil
	})

	if err != nil {
		return value.PaymentRegistrationRequestParam{}, nil
	}

	requestParam, err := u.service.userPayment.GetCardRegistrationRequestParam(ctx, user)
	if err != nil {
		return value.PaymentRegistrationRequestParam{}, fmt.Errorf("app.user.GetCardRegistrationRequestParam: error from payment service: %w", err)
	}

	return requestParam, nil
}

func (u userApp) RegistrationCallback(ctx context.Context, req request.PaymentRegistrationCallbackRequest) (entity.UserPayment, error) {
	userPayment, err := u.service.userPayment.RegistrationCallback(ctx, req)
	if err != nil {
		return entity.UserPayment{}, fmt.Errorf("app.user.RegistrationCallback: error from payment service: %w", err)
	}

	return userPayment, nil
}

func (u userApp) ListUserPayment(ctx context.Context, userId string) ([]entity.UserPayment, error) {
	payments, err := u.service.userPayment.ListUserPayment(ctx, userId)
	if err != nil {
		return []entity.UserPayment{}, fmt.Errorf("app.user.ListUserPayment: error while list user payments from payment app: %w", err)
	}
	return payments, nil
}

func (u userApp) TryRecoverUserPayment(ctx context.Context, userPaymentId string) error {
	userId := utils.GetUserId(ctx)
	err := u.service.userPayment.TryRecoverUserPayment(ctx, userId, userPaymentId)
	if err != nil {
		return fmt.Errorf("app.user.TryRecoverUserPayment: error from payment service: %w", err)
	}
	return nil
}

func (u userApp) DeleteUserPayment(ctx context.Context, userPaymentId string) error {
	userId := utils.GetUserId(ctx)

	var user entity.User
	err := u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		u, err := u.repository.user.FindById(ctx, i, userId)
		if err != nil {
			return fmt.Errorf("app.user.DeleteUserPayment: Error while find user by id:%w", err)
		}
		user = u
		return nil
	})
	if err != nil {
		return err
	}

	err = u.service.userPayment.DeleteUserPayment(ctx, user, userPaymentId)
	if err != nil {
		return fmt.Errorf("app.user.DeleteUserPayment: error while delete user payment: %w", err)
	}

	return nil
}

func (u userApp) PaymentTransactionSuccessCallback(ctx context.Context, req request.PaymentTransactionSuccessCallbackRequest) error {
	return u.service.userPayment.PaymentTransactionSuccessCallback(ctx, req)
}

func (u userApp) PaymentTransactionFailCallback(ctx context.Context, req request.PaymentTransactionFailCallbackRequest) error {
	return u.service.userPayment.PaymentTransactionFailCallback(ctx, req)
}
