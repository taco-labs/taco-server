package payment

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/event/command"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/domain/value/analytics"
	"github.com/taco-labs/taco/go/repository"
	"github.com/taco-labs/taco/go/service"
	"github.com/taco-labs/taco/go/utils"
	"github.com/taco-labs/taco/go/utils/slices"
	"github.com/uptrace/bun"
	"go.uber.org/zap"
)

type paymentApp struct {
	app.Transactor

	repository struct {
		payment   repository.PaymentRepository
		referral  repository.ReferralRepository
		event     repository.EventRepository
		analytics repository.AnalyticsRepository
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
	logger := utils.GetLogger(ctx)

	var userPayment entity.UserPayment
	var paymentRegistrationRequest entity.UserPaymentRegistrationRequest

	err := u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		registrationRequest, err := u.repository.payment.GetUserPaymentRegistrationRequest(ctx, i, req.RequestId)
		if errors.Is(err, value.ErrNotFound) {
			// TODO (taekyeom) already reigstered... just logging with warn later
			return nil
		}
		if err != nil {
			return fmt.Errorf("app.payment.RegistrationCallback: error while get payment registration request")
		}

		paymentRegistrationRequest = registrationRequest

		userPayment = entity.UserPayment{
			Id:                 registrationRequest.PaymentId,
			UserId:             registrationRequest.UserId,
			CardCompany:        req.CardCompany,
			RedactedCardNumber: req.CardNumber,
			BillingKey:         req.BillingKey,
			Invalid:            false,
			CreateTime:         requestTime,
			LastUseTime:        time.Time{},
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
			orderCommands := slices.Map(failedOrders, func(i entity.UserPaymentFailedOrder) entity.Event {
				return command.NewUserPaymentTransactionRequestCommand(
					registrationRequest.UserId,
					registrationRequest.PaymentId,
					i.OrderId,
					i.OrderName,
					i.SettlementTargetId,
					i.Amount,
					i.UsedPoint,
					i.SettlementAmount,
					i.AdditionalSettlementAmount,
					false,
				)
			})
			if err := u.repository.event.BatchCreate(ctx, i, orderCommands); err != nil {
				return fmt.Errorf("app.payment.RegistrationCallback: Error while create recovery payment events: %w", err)
			}
		}

		if err := u.repository.payment.DeleteUserPaymentRegistrationRequest(ctx, i, registrationRequest); err != nil {
			return fmt.Errorf("app.payment.RegistrationCallback: Error while delete registration request: %w", err)
		}

		return nil
	})

	if err != nil && paymentRegistrationRequest.PaymentId != "" {
		u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
			cmd := command.NewPaymentUserPaymentDeleteCommand(
				paymentRegistrationRequest.UserId,
				paymentRegistrationRequest.PaymentId,
				req.BillingKey,
			)
			if err := u.repository.event.BatchCreate(ctx, i, []entity.Event{cmd}); err != nil {
				logger.Error("app.payment.RegistrationCallback: error while publish card delete command when card registration failed",
					zap.Error(err),
				)
			}
			return nil
		})
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

func (u paymentApp) UpdateUserPayment(ctx context.Context, userPayment entity.UserPayment) error {
	return u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		if err := u.repository.payment.UpdateUserPayment(ctx, i, userPayment); err != nil {
			return fmt.Errorf("app.paymentApp.UpdateUserPayment: error while update user payment: %w", err)
		}

		return nil
	})
}

func (u paymentApp) ListUserPayment(ctx context.Context, userId string) ([]entity.UserPayment, error) {
	var userPayments []entity.UserPayment
	var err error

	err = u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		userPayments, err = u.repository.payment.ListUserPayment(ctx, i, userId)
		if err != nil {
			return fmt.Errorf("app.userPayment.ListUserPayment: Error while list payments:%w", err)
		}

		return nil
	})

	if err != nil {
		return []entity.UserPayment{}, err
	}

	return userPayments, nil
}

func (p paymentApp) TryRecoverUserPayment(ctx context.Context, userId string, userPaymentId string) error {
	return p.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		userPayment, err := p.repository.payment.GetUserPayment(ctx, i, userPaymentId)
		if err != nil {
			return fmt.Errorf("app.userPayment.TryRecoverUserPayment: error while get user payment: %w", err)
		}
		if userPayment.UserId != userId {
			return fmt.Errorf("app.userPayment.TryRecoverUserPayment: unauthorized payment: %w", value.ErrUnAuthorized)
		}

		failedOrders, err := p.repository.payment.GetFailedOrdersByUserId(ctx, i, userId)
		if err != nil {
			return fmt.Errorf("app.payment.handleTransactionSuccess: failed to list failed order: %w", err)
		}

		if len(failedOrders) > 0 {
			recoveryCommand := command.NewUserPaymentTransactionRequestCommand(
				userId,
				userPaymentId,
				failedOrders[0].OrderId,
				failedOrders[0].OrderName,
				failedOrders[0].SettlementTargetId,
				failedOrders[0].Amount,
				failedOrders[0].UsedPoint,
				failedOrders[0].SettlementAmount,
				failedOrders[0].AdditionalSettlementAmount,
				true,
			)
			if err := p.repository.event.BatchCreate(ctx, i, []entity.Event{recoveryCommand}); err != nil {
				return fmt.Errorf("app.userPayment.TryRecoverUserPayment: Error while create recovery payment events: %w", err)
			}
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

func (p paymentApp) PaymentTransactionSuccessCallback(ctx context.Context, req request.PaymentTransactionSuccessCallbackRequest) error {
	return p.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		cmd := command.NewUserPaymentTransactionSuccessCommand(
			req.OrderId,
			req.PaymentKey,
			req.ReceiptUrl,
			req.CreateTime,
		)
		if err := p.repository.event.BatchCreate(ctx, i, []entity.Event{cmd}); err != nil {
			return fmt.Errorf("app.paymentApp.PaymentTransactionSuccessCallback: error while create success event: %w", err)
		}
		return nil
	})
}

func (p paymentApp) PaymentTransactionFailCallback(ctx context.Context, req request.PaymentTransactionFailCallbackRequest) error {
	return p.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		cmd := command.NewUserPaymentTransactionFailCommand(
			req.OrderId,
			req.FailureCode,
			req.FailureReason,
		)
		if err := p.repository.event.BatchCreate(ctx, i, []entity.Event{cmd}); err != nil {
			return fmt.Errorf("app.paymentApp.PaymentTransactionFailCallback: error while create fail event: %w", err)
		}
		return nil
	})
}

func (p paymentApp) CreateUserPaymentPoint(ctx context.Context, userPaymentPoint entity.UserPaymentPoint) error {
	return p.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		if err := p.repository.payment.CreateUserPaymentPoint(ctx, i, userPaymentPoint); err != nil {
			return fmt.Errorf("app.paymentApp.CreateUserPaymentPoint: error while create user payment point: %w", err)
		}

		return nil
	})
}

func (p paymentApp) GetUserPaymentPoint(ctx context.Context, userId string) (entity.UserPaymentPoint, error) {
	var userPaymentPoint entity.UserPaymentPoint
	err := p.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		up, err := p.repository.payment.GetUserPaymentPoint(ctx, i, userId)
		if err != nil {
			return fmt.Errorf("app.paymentApp.GetUserPaymentPoint: error while get user payment point: %w", err)
		}

		userPaymentPoint = up

		return nil
	})

	if err != nil {
		return entity.UserPaymentPoint{}, err
	}

	return userPaymentPoint, nil
}

func (p paymentApp) UseUserPaymentPoint(ctx context.Context, userId string, price int) (int, error) {
	// short circuit
	if price == 0 {
		return 0, nil
	}

	var usedPoint int
	err := p.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		userPaymentPoint, err := p.repository.payment.GetUserPaymentPoint(ctx, i, userId)
		if err != nil {
			return fmt.Errorf("app.paymentApp.UseUserPaymentPoint: error while get user payment point: %w", err)
		}
		usedPoint = userPaymentPoint.UsePoint(price)

		if usedPoint > 0 {
			if err := p.repository.payment.UpdateUserPaymentPoint(ctx, i, userPaymentPoint); err != nil {
				return fmt.Errorf("app.paymentApp.UseUserPaymentPoint: error while update user payment point: %w", err)
			}
		}
		return nil
	})

	if err != nil {
		return 0, err
	}
	return usedPoint, nil
}

func (p paymentApp) AddUserPaymentPoint(ctx context.Context, userId string, point int) error {
	if point == 0 {
		return nil
	}
	return p.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		userPaymentPoint, err := p.repository.payment.GetUserPaymentPoint(ctx, i, userId)
		if err != nil {
			return fmt.Errorf("app.paymentApp.CancelUserPaymentPoint: error while get user payment point: %w", err)
		}
		userPaymentPoint.AddPoint(point)

		if err := p.repository.payment.UpdateUserPaymentPoint(ctx, i, userPaymentPoint); err != nil {
			return fmt.Errorf("app.paymentApp.CancelUserPaymentPoint: error while update user payment point: %w", err)
		}
		return nil
	})
}

func (p paymentApp) CreateUserReferral(ctx context.Context, fromUserId string, toUserId string) error {
	requestTime := utils.GetRequestTimeOrNow(ctx)
	return p.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		userReferral := entity.UserReferral{
			FromUserId:    fromUserId,
			ToUserId:      toUserId,
			CurrentReward: 0,
			RewardLimit:   entity.UserReferralRewardLimit,
			RewardRate:    entity.UserReferralRewardRate,
			CreateTime:    requestTime,
		}

		if err := p.repository.referral.CreateUserReferral(ctx, i, userReferral); err != nil {
			return fmt.Errorf("app.paymentApp.CreateUserReferral: error while create user referral: %w", err)
		}

		return nil
	})
}

func (p paymentApp) ApplyUserReferralReward(ctx context.Context, userId, orderId string, price int) error {
	requestTime := utils.GetRequestTimeOrNow(ctx)
	return p.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		userReferralReward, err := p.repository.referral.GetUserReferral(ctx, i, userId)
		if errors.Is(err, value.ErrNotFound) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("app.paymentApp.UseUserReferralReward: error while get user referral reward: %w", err)
		}
		reward := userReferralReward.UseReward(price)
		if err := p.repository.referral.UpdateUserReferral(ctx, i, userReferralReward); err != nil {
			return fmt.Errorf("app.paymentApp.UseUserReferralReward: error while update user referral reward: %w", err)
		}

		userPaymentPoint, err := p.repository.payment.GetUserPaymentPoint(ctx, i, userReferralReward.ToUserId)
		if err != nil {
			return fmt.Errorf("app.paymentApp.UseUserReferralReward: error while get user payment point: %w", err)
		}
		userPaymentPoint.Point += reward
		if err := p.repository.payment.UpdateUserPaymentPoint(ctx, i, userPaymentPoint); err != nil {
			return fmt.Errorf("app.paymentApp.UseUserReferralReward: error while update user payment point: %w", err)
		}

		referralRewardNotification := NewReferralRewardNotification(userReferralReward.ToUserId, reward)
		if err := p.repository.event.BatchCreate(ctx, i, []entity.Event{referralRewardNotification}); err != nil {
			return fmt.Errorf("app.paymentApp.UseUserReferralReward: error while add notification event: %w", err)
		}

		referralRewardAnalytics := entity.NewAnalytics(requestTime, analytics.UserReferralPointReceivedPayload{
			UserId:        userReferralReward.ToUserId,
			FromUserId:    userReferralReward.ToUserId,
			OrderId:       orderId,
			ReceiveAmount: reward,
		})
		if err := p.repository.analytics.Create(ctx, i, referralRewardAnalytics); err != nil {
			return fmt.Errorf("app.paymanetApp.ApplyUserReferralReward: error while create referral reward analytics event: %w", err)
		}

		return nil
	})
}

func (p paymentApp) CreateDriverReferral(ctx context.Context, driverId string) error {
	requestTime := utils.GetRequestTimeOrNow(ctx)
	return p.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		driverReferral := entity.DriverReferral{
			DriverId:      driverId,
			CurrentReward: 0,
			RewardLimit:   0,
			RewardRate:    entity.DriverReferralRewardRate,
			CreateTime:    requestTime,
		}

		if err := p.repository.referral.CreateDriverReferral(ctx, i, driverReferral); err != nil {
			return fmt.Errorf("app.paymentApp.CreateUserReferral: error while create user referral: %w", err)
		}

		return nil
	})
}

func (p paymentApp) AddDriverReferralReward(ctx context.Context, driverId string) error {
	return p.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		driverReferral, err := p.repository.referral.GetDriverReferral(ctx, i, driverId)
		if err != nil {
			return fmt.Errorf("app.paymentApp.AddDriverReferralReward: error while get driver referral: %w", err)
		}

		driverReferral.RewardLimit += entity.DriverReferralRewardLimit
		if err := p.repository.referral.UpdateDriverReferral(ctx, i, driverReferral); err != nil {
			return fmt.Errorf("app.paymentApp.AddDriverReferralReward: error while update driver referra: %w", err)
		}

		return nil
	})
}

func (p paymentApp) UseDriverReferralReward(ctx context.Context, driverId string, price int) (int, error) {
	var reward int
	err := p.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		driverReferral, err := p.repository.referral.GetDriverReferral(ctx, i, driverId)
		if err != nil {
			return fmt.Errorf("app.paymentApp.UseDriverReferralReward: error while get driver referral: %w", err)
		}

		reward = driverReferral.UseReweard(price)

		if err := p.repository.referral.UpdateDriverReferral(ctx, i, driverReferral); err != nil {
			return fmt.Errorf("app.paymentApp.UseDriverReferralReward: error while update driver referra: %w", err)
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return reward, nil
}

func (p paymentApp) CancelDriverReferralReward(ctx context.Context, driverId string, cancelReward int) error {
	return p.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		driverReferral, err := p.repository.referral.GetDriverReferral(ctx, i, driverId)
		if err != nil {
			return fmt.Errorf("app.paymentApp.CancelDriverReferralReward: error while get driver referral: %w", err)
		}

		driverReferral.CancelReward(cancelReward)
		if err := p.repository.referral.UpdateDriverReferral(ctx, i, driverReferral); err != nil {
			return fmt.Errorf("app.paymentApp.CancelDriverReferralReward: error while update driver referra: %w", err)
		}

		return nil
	})
}
