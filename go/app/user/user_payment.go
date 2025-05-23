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

// TODO(taekyeom) Format error

func (u userApp) ListCardPayment(ctx context.Context, userId string) ([]entity.UserPayment, entity.UserDefaultPayment, error) {
	var userPayments []entity.UserPayment
	var userDefaultPayment entity.UserDefaultPayment
	var err error

	err = u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		userPayments, err = u.repository.payment.ListUserPayment(ctx, i, userId)
		if err != nil {
			return fmt.Errorf("app.user.ListCardPayment: Error while list payments:%w", err)
		}

		userDefaultPayment, err = u.repository.payment.GetDefaultPaymentByUserId(ctx, i, userId)
		if err != nil {
			return fmt.Errorf("app.user.ListCardPayment: Error while get default payment payments:%w", err)
		}

		return nil
	})

	if err != nil {
		return []entity.UserPayment{}, entity.UserDefaultPayment{}, err
	}

	return userPayments, userDefaultPayment, nil
}

func (u userApp) RegisterCardPayment(ctx context.Context, req request.UserPaymentRegisterRequest) (entity.UserPayment, error) {
	userId := utils.GetUserId(ctx)
	requestTime := utils.GetRequestTimeOrNow(ctx)

	payment, err := u.service.payment.RegisterCard(ctx, utils.MustNewUUID(), req)
	if err != nil {
		return entity.UserPayment{}, fmt.Errorf("app.user.RegisterCardPayment: Error while register card: %w", err)
	}

	var userPayment entity.UserPayment
	err = u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		user, err := u.repository.user.FindById(ctx, i, userId)
		if err != nil {
			return fmt.Errorf("app.user.RegisterCardPayment: Error while find user by id:%w", err)
		}

		userPayment = entity.UserPayment{
			Id:                  payment.CustomerKey,
			UserId:              user.Id,
			Name:                req.Name,
			CardCompany:         payment.CardCompany,
			RedactedCardNumber:  payment.CardNumber,
			CardExpirationYear:  payment.CardExpirationYear,
			CardExpirationMonth: payment.CardExpirationMonth,
			BillingKey:          payment.BillingKey,
			DefaultPayment:      req.DefaultPayment,
			CreateTime:          requestTime,
		}

		err = u.repository.payment.CreateUserPayment(ctx, i, userPayment)
		if err != nil {
			return fmt.Errorf("app.user.RegisterCardPayment: Error while create user payment:%w", err)
		}

		if userPayment.DefaultPayment {
			userDefaultPayment := entity.UserDefaultPayment{
				UserId:    userPayment.UserId,
				PaymentId: userPayment.Id,
			}
			if err = u.repository.payment.UpsertDefaultPayment(ctx, i, userDefaultPayment); err != nil {
				return fmt.Errorf("app.user.RegisterCardPayment: Error while update default payment: %w", err)
			}
		}
		return nil
	})

	if err != nil {
		return entity.UserPayment{}, err
	}

	return userPayment, nil
}

func (u userApp) DeleteCardPayment(ctx context.Context, userPaymentId string) error {
	userId := utils.GetUserId(ctx)

	return u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		user, err := u.repository.user.FindById(ctx, i, userId)
		if err != nil {
			return fmt.Errorf("app.user.ListCardPayment: Error while find by user id:%w", err)
		}

		userPayment, err := u.repository.payment.GetUserPayment(ctx, i, userPaymentId)
		if err != nil {
			return fmt.Errorf("app.user.ListCardPayment: Error while find user payment:%w", err)
		}

		if user.Id != userPayment.UserId {
			return fmt.Errorf("app.user.ListCardPayment: user id does not matched:%w", value.ErrUnAuthorized)
		}

		err = u.repository.payment.DeleteUserPayment(ctx, i, userPaymentId)
		if err != nil {
			return err
		}

		return nil
	})
}

func (u userApp) UpdateDefaultPayment(ctx context.Context, req request.DefaultPaymentUpdateRequest) error {
	return u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		userDefaultPayment, err := u.repository.payment.GetDefaultPaymentByUserId(ctx, i, req.Id)
		if err != nil {
			return fmt.Errorf("app.user.UpdateDefaultPayment: error while find default user default payment by user id:\n %w", err)
		}

		if userDefaultPayment.PaymentId == req.PaymentId {
			return nil
		}

		userPayment, err := u.repository.payment.GetUserPayment(ctx, i, req.PaymentId)
		if err != nil {
			return err
		}

		if req.Id != userPayment.UserId {
			return value.ErrUnAuthorized
		}

		userDefaultPayment.PaymentId = userPayment.Id

		if err = u.repository.payment.UpsertDefaultPayment(ctx, i, userDefaultPayment); err != nil {
			return fmt.Errorf("app.user.UpdateDefaultPayment: error while update default payment:\n %w", err)
		}

		return nil
	})
}
