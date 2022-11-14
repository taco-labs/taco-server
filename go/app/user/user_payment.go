package user

import (
	"context"
	"fmt"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/utils"
	"github.com/uptrace/bun"
)

func (u userApp) ListUserPayment(ctx context.Context, userId string) ([]entity.UserPayment, entity.UserDefaultPayment, error) {
	payments, defaultPayment, err := u.service.userPayment.ListUserPayment(ctx, userId)
	if err != nil {
		return []entity.UserPayment{}, entity.UserDefaultPayment{}, fmt.Errorf("app.user.ListUserPayment: error while list user payments from payment app: %w", err)
	}
	return payments, defaultPayment, nil
}

func (u userApp) RegisterUserPayment(ctx context.Context, req request.UserPaymentRegisterRequest) (entity.UserPayment, error) {
	userId := utils.GetUserId(ctx)

	var user entity.User
	err := u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		u, err := u.repository.user.FindById(ctx, i, userId)
		if err != nil {
			return fmt.Errorf("app.user.RegisterUserPayment: Error while find user by id:%w", err)
		}
		user = u
		return nil
	})
	if err != nil {
		return entity.UserPayment{}, err
	}

	userPayment, err := u.service.userPayment.RegisterUserPayment(ctx, user, req)
	if err != nil {
		return entity.UserPayment{}, fmt.Errorf("app.user.RegisterUserPayment: error while register user payment: %w", err)
	}

	return userPayment, nil
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

func (u userApp) UpdateDefaultPayment(ctx context.Context, req request.DefaultPaymentUpdateRequest) error {
	err := u.service.userPayment.UpdateDefaultPayment(ctx, req)
	if err != nil {
		return fmt.Errorf("app.user.UpdateDefaultPayment: error while update default payment: %w", err)
	}

	return nil
}
