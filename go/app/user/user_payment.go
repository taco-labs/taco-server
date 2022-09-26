package user

import (
	"context"
	"fmt"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/utils"
)

// TODO(taekyeom) Format error

func (u userApp) ListCardPayment(ctx context.Context, userId string) ([]entity.UserPayment, entity.UserDefaultPayment, error) {
	ctx, err := u.Start(ctx)
	if err != nil {
		return []entity.UserPayment{}, entity.UserDefaultPayment{}, err
	}
	defer func() {
		err = u.Done(ctx, err)
	}()

	userPayments, err := u.repository.payment.ListUserPayment(ctx, userId)
	if err != nil {
		return []entity.UserPayment{}, entity.UserDefaultPayment{}, err
	}

	userDefaultPayment, err := u.repository.payment.GetDefaultPaymentByUserId(ctx, userId)
	if err != nil {
		return []entity.UserPayment{}, entity.UserDefaultPayment{}, err
	}

	return userPayments, userDefaultPayment, nil
}

func (u userApp) RegisterCardPayment(ctx context.Context, req request.UserPaymentRegisterRequest) (entity.UserPayment, error) {
	ctx, err := u.Start(ctx)
	if err != nil {
		return entity.UserPayment{}, err
	}
	defer func() {
		err = u.Done(ctx, err)
	}()

	userId := utils.GetUserId(ctx)

	user, err := u.repository.user.FindById(ctx, userId)
	if err != nil {
		return entity.UserPayment{}, err
	}

	userPayment, err := u.service.payment.RegisterCard(ctx, user, req)
	if err != nil {
		return entity.UserPayment{}, err
	}

	err = u.repository.payment.CreateUserPayment(ctx, userPayment)
	if err != nil {
		return entity.UserPayment{}, err
	}

	if req.DefaultPayment {
		userDefaultPayment := entity.UserDefaultPayment{
			UserId:    userId,
			PaymentId: userPayment.Id,
		}
		if err = u.repository.payment.UpsertDefaultPayment(ctx, userDefaultPayment); err != nil {
			return entity.UserPayment{}, err
		}
	}

	return userPayment, nil
}

func (u userApp) DeleteCardPayment(ctx context.Context, userPaymentId string) error {
	ctx, err := u.Start(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = u.Done(ctx, err)
	}()

	userId := utils.GetUserId(ctx)

	user, err := u.repository.user.FindById(ctx, userId)
	if err != nil {
		return err
	}

	userPayment, err := u.repository.payment.GetUserPayment(ctx, userPaymentId)
	if err != nil {
		return err
	}

	if user.Id != userPayment.UserId {
		return value.ErrUnAuthorized
	}

	err = u.repository.payment.DeleteUserPayment(ctx, userPaymentId)
	if err != nil {
		return err
	}

	return nil
}

func (u userApp) UpdateDefaultPayment(ctx context.Context, req request.DefaultPaymentUpdateRequest) error {
	ctx, err := u.Start(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = u.Done(ctx, err)
	}()

	userDefaultPayment, err := u.repository.payment.GetDefaultPaymentByUserId(ctx, req.Id)
	if err != nil {
		return fmt.Errorf("app.user.UpdateDefaultPayment: error while find default user default payment by user id:\n %w", err)
	}

	if userDefaultPayment.PaymentId == req.PaymentId {
		return nil
	}

	userPayment, err := u.repository.payment.GetUserPayment(ctx, req.PaymentId)
	if err != nil {
		return err
	}

	if req.Id != userPayment.UserId {
		return value.ErrUnAuthorized
	}

	userDefaultPayment.PaymentId = userPayment.Id

	if err = u.repository.payment.UpsertDefaultPayment(ctx, userDefaultPayment); err != nil {
		return fmt.Errorf("app.user.UpdateDefaultPayment: error while update default payment:\n %w", err)
	}

	return nil
}
