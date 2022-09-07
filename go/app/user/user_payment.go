package user

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ktk1012/taco/go/domain/entity"
	"github.com/ktk1012/taco/go/domain/request"
	"github.com/ktk1012/taco/go/domain/value"
	"github.com/ktk1012/taco/go/utils"
)

// TODO(taekyeom) Format error

func (u userApp) ListCardPayment(ctx context.Context, userId string) ([]entity.UserPayment, error) {
	userPayments, err := u.repository.payment.ListUserPayment(ctx, userId)
	if err != nil {
		return []entity.UserPayment{}, err
	}

	return userPayments, nil
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
		user.DefaultPaymentId = sql.NullString{
			String: userPayment.Id,
			Valid:  true,
		}
		if err = u.repository.user.UpdateUser(ctx, user); err != nil {
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

func (u userApp) UpdateDefaultPayment(ctx context.Context, req request.DefaultPaymentUpdateRequest) (entity.User, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	ctx, err := u.Start(ctx)
	if err != nil {
		return entity.User{}, err
	}
	defer func() {
		err = u.Done(ctx, err)
	}()

	user, err := u.repository.user.FindById(ctx, req.Id)
	if err != nil {
		return entity.User{}, fmt.Errorf("app.user.UpdateDefaultPayment: error while find user by id:\n %v", err)
	}

	if user.DefaultPaymentId.String == req.DefaultPaymentId {
		return user, nil
	}

	userPayment, err := u.repository.payment.GetUserPayment(ctx, req.DefaultPaymentId)
	if err != nil {
		return entity.User{}, err
	}

	if user.Id != userPayment.UserId {
		return entity.User{}, value.ErrUnAuthorized
	}

	user.DefaultPaymentId = sql.NullString{
		String: req.DefaultPaymentId,
		Valid:  true,
	}
	user.UpdateTime = requestTime

	if err = u.repository.user.UpdateUser(ctx, user); err != nil {
		return entity.User{}, fmt.Errorf("app.user.UpdateDefaultPayment: error while update default payment:\n %v", err)
	}

	return user, nil
}
