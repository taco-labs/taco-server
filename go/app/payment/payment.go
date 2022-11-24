package payment

import (
	"context"
	"errors"
	"fmt"

	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/event/command"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/repository"
	"github.com/taco-labs/taco/go/service"
	"github.com/taco-labs/taco/go/utils"
	"github.com/taco-labs/taco/go/utils/slices"
	"github.com/uptrace/bun"
)

type paymentApp struct {
	app.Transactor

	repository struct {
		payment repository.PaymentRepository
		event   repository.EventRepository
	}

	service struct {
		payment service.PaymentService
	}
}

func (u paymentApp) GetCardRegistrationRequestParam(ctx context.Context, user entity.User) (value.PaymentRegistrationRequestParam, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	var requestParam value.PaymentRegistrationRequestParam
	err := u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		registrationRequest, err := u.repository.payment.CreateUserPaymentRegistrationRequest(ctx, i, entity.UserPaymentRegistrationRequest{
			PaymentId:  utils.MustNewUUID(),
			UserId:     user.Id,
			CreateTime: requestTime,
		})
		if err != nil {
			return fmt.Errorf("app.payment.GetCardRegistrationRequestParam: error while create request entity: %w", err)
		}

		rp, err := u.service.payment.GetCardRegistrationRequestParam(ctx, registrationRequest.RequestId, user)
		if err != nil {
			return fmt.Errorf("app.payment.GetCardRegistrationRequestParam: error while get requets param: %w", err)
		}

		requestParam = rp

		return nil
	})

	if err != nil {
		return value.PaymentRegistrationRequestParam{}, err
	}

	return requestParam, nil
}

func (u paymentApp) RegistrationCallback(ctx context.Context, req request.PaymentRegistrationCallbackRequest) (entity.UserPayment, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	var userPayment entity.UserPayment

	err := u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		registrationRequest, err := u.repository.payment.GetUserPaymentRegistrationRequest(ctx, i, req.RequestId)
		if errors.Is(err, value.ErrNotFound) {
			// TODO (taekyeom) already reigstered... just logging with warn later
			return nil
		}
		if err != nil {
			return fmt.Errorf("app.payment.RegistrationCallback: error while get payment registration request")
		}

		userPayment = entity.UserPayment{
			Id:                 registrationRequest.PaymentId,
			UserId:             registrationRequest.UserId,
			CardCompany:        req.CardCompany,
			RedactedCardNumber: req.CardNumber,
			BillingKey:         req.BillingKey,
			Invalid:            false,
			CreateTime:         requestTime,
		}

		if err := u.repository.payment.CreateUserPayment(ctx, i, userPayment); err != nil {
			return fmt.Errorf("app.payment.RegistrationCallback: error while create user payment: %w", err)
		}

		// TODO (taekyeom) publish event & handle in background?
		failedOrders, err := u.repository.payment.GetFailedOrdersByUserId(ctx, i, userPayment.UserId)
		if err != nil {
			return fmt.Errorf("app.payment.RegistrationCallback: Error while get failed orders: %w", err)
		}
		if len(failedOrders) > 0 {
			recoveryCommand := command.NewPaymentUserTransactionRecoveryCommand(userPayment.UserId, userPayment.Id)
			if err := u.repository.event.BatchCreate(ctx, i, []entity.Event{recoveryCommand}); err != nil {
				return fmt.Errorf("app.payment.RegistrationCallback: Error while create recovery payment events: %w", err)
			}
		}

		if err := u.repository.payment.DeleteUserPaymentRegistrationRequest(ctx, i, registrationRequest); err != nil {
			return fmt.Errorf("app.payment.RegistrationCallback: Error while delete registration request: %w", err)
		}

		return nil
	})

	if err != nil {
		return entity.UserPayment{}, err
	}

	return userPayment, nil
}

func (u paymentApp) GetUserPayment(ctx context.Context, userId string, userPaymentId string) (entity.UserPayment, error) {
	var userPayment entity.UserPayment
	err := u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		p, err := u.repository.payment.GetUserPayment(ctx, i, userPaymentId)
		if err != nil {
			return fmt.Errorf("app.userPayment.GetUserPayment: error while get user payment: %w", err)
		}
		if p.UserId != userId {
			return fmt.Errorf("app.userPayment.GetUserPayment: unauthorized payment: %w", value.ErrUnAuthorized)
		}
		userPayment = p
		return nil
	})

	if err != nil {
		return entity.UserPayment{}, err
	}

	if userPayment.Invalid {
		err = value.NewTacoError(value.ERR_INVALID_USER_PAYMENT, userPayment.InvalidErrorMessage)
		return entity.UserPayment{}, fmt.Errorf("app.userPayment.GetUserPayment: invalid user payment: %w", err)
	}

	return userPayment, nil
}

func (u paymentApp) ListUserPayment(ctx context.Context, userId string) ([]entity.UserPayment, entity.UserDefaultPayment, error) {
	var userPayments []entity.UserPayment
	var userDefaultPayment entity.UserDefaultPayment
	var err error

	err = u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		userPayments, err = u.repository.payment.ListUserPayment(ctx, i, userId)
		if err != nil {
			return fmt.Errorf("app.userPayment.ListUserPayment: Error while list payments:%w", err)
		}

		userDefaultPayment, err = u.repository.payment.GetDefaultPaymentByUserId(ctx, i, userId)
		if err != nil {
			return fmt.Errorf("app.userPayment.ListUserPayment: Error while get default payment payments:%w", err)
		}

		return nil
	})

	if err != nil {
		return []entity.UserPayment{}, entity.UserDefaultPayment{}, err
	}

	return userPayments, userDefaultPayment, nil
}

func (u paymentApp) RegisterUserPayment(ctx context.Context, user entity.User, req request.UserPaymentRegisterRequest) (entity.UserPayment, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	payment, err := u.service.payment.RegisterCard(ctx, utils.MustNewUUID(), req)
	if err != nil {
		return entity.UserPayment{}, fmt.Errorf("app.userPayment.RegisterUserPayment: Error while register card: %w", err)
	}

	var userPayment entity.UserPayment
	err = u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
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
			return fmt.Errorf("app.userPayment.RegisterUserPayment: Error while create user payment:%w", err)
		}

		if userPayment.DefaultPayment {
			userDefaultPayment := entity.UserDefaultPayment{
				UserId:    userPayment.UserId,
				PaymentId: userPayment.Id,
			}
			if err = u.repository.payment.UpsertDefaultPayment(ctx, i, userDefaultPayment); err != nil {
				return fmt.Errorf("app.userPayment.RegisterUserPayment: Error while update default payment: %w", err)
			}
		}

		// TODO (taekyeom) publish event & handle in background?
		failedOrders, err := u.repository.payment.GetFailedOrdersByUserId(ctx, i, user.Id)
		if err != nil {
			return fmt.Errorf("app.userPayment.RegisterUserPayment: Error while get failed orders: %w", err)
		}
		if len(failedOrders) > 0 {
			recoveryCommand := command.NewPaymentUserTransactionRecoveryCommand(user.Id, userPayment.Id)
			if err := u.repository.event.BatchCreate(ctx, i, []entity.Event{recoveryCommand}); err != nil {
				return fmt.Errorf("app.userPayment.TryRecoverUserPayment: Error while create recovery payment events: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return entity.UserPayment{}, err
	}

	// TODO (taekyeom) Handle failed order when new card is registered

	return userPayment, nil
}

func (u paymentApp) TryRecoverUserPayment(ctx context.Context, userId string, userPaymentId string) error {
	return u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		userPayment, err := u.repository.payment.GetUserPayment(ctx, i, userPaymentId)
		if err != nil {
			return fmt.Errorf("app.userPayment.TryRecoverUserPayment: error while get user payment: %w", err)
		}
		if userPayment.UserId != userId {
			return fmt.Errorf("app.userPayment.TryRecoverUserPayment: unauthorized payment: %w", value.ErrUnAuthorized)
		}

		recoveryCommand := command.NewPaymentUserTransactionRecoveryCommand(userId, userPaymentId)
		if err := u.repository.event.BatchCreate(ctx, i, []entity.Event{recoveryCommand}); err != nil {
			return fmt.Errorf("app.userPayment.TryRecoverUserPayment: Error while create recovery payment events: %w", err)
		}

		return nil
	})
}

func (u paymentApp) DeleteUserPayment(ctx context.Context, user entity.User, userPaymentId string) error {
	return u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		userPayment, err := u.repository.payment.GetUserPayment(ctx, i, userPaymentId)
		if err != nil {
			return fmt.Errorf("app.userPayment.DeleteUserPayment: Error while find user payment:%w", err)
		}

		if user.Id != userPayment.UserId {
			return fmt.Errorf("app.userPayment.DeleteUserPayment: user id does not matched:%w", value.ErrUnAuthorized)
		}

		if err = u.repository.payment.DeleteUserPayment(ctx, i, userPaymentId); err != nil {
			return fmt.Errorf("app.userPayment.DeleteUserPayment: error while delete user payment: %w", err)
		}

		deleteCmd := command.NewPaymentUserPaymentDeleteCommand(user.Id, userPaymentId, userPayment.BillingKey)
		if err = u.repository.event.BatchCreate(ctx, i, []entity.Event{deleteCmd}); err != nil {
			return fmt.Errorf("app.userPayment.DeleteUserPayment: error while create user payment delete command: %w", err)
		}

		return nil
	})
}

func (u paymentApp) BatchDeleteUserPayment(ctx context.Context, user entity.User) error {
	return u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		deletedUserPayments, err := u.repository.payment.BatchDeleteUserPayment(ctx, i, user.Id)
		if err != nil {
			return fmt.Errorf("app.userPayment.BatchDeleteUserPayment: error while batch delete user payment: %w", err)
		}

		deleteCmds := slices.Map(deletedUserPayments, func(i entity.UserPayment) entity.Event {
			return command.NewPaymentUserPaymentDeleteCommand(i.UserId, i.Id, i.BillingKey)
		})

		if err := u.repository.event.BatchCreate(ctx, i, deleteCmds); err != nil {
			return fmt.Errorf("app.userPayment.BatchDeleteUserPayment: error while create payment delete commands: %w", err)
		}

		return nil
	})
}

func (u paymentApp) UpdateDefaultPayment(ctx context.Context, req request.DefaultPaymentUpdateRequest) error {
	return u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		userDefaultPayment, err := u.repository.payment.GetDefaultPaymentByUserId(ctx, i, req.UserId)
		if err != nil {
			return fmt.Errorf("app.userPayment.UpdateDefaultPayment: error while find default user default payment by user id:\n %w", err)
		}

		if userDefaultPayment.PaymentId == req.PaymentId {
			return nil
		}

		userPayment, err := u.repository.payment.GetUserPayment(ctx, i, req.PaymentId)
		if err != nil {
			return err
		}

		if req.UserId != userPayment.UserId {
			return value.ErrUnAuthorized
		}

		userDefaultPayment.PaymentId = userPayment.Id

		if err = u.repository.payment.UpsertDefaultPayment(ctx, i, userDefaultPayment); err != nil {
			return fmt.Errorf("app.userPayment.UpdateDefaultPayment: error while update default payment:\n %w", err)
		}

		return nil
	})
}
