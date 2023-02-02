package payment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/event/command"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/domain/value/enum"
	"github.com/taco-labs/taco/go/utils"
	"github.com/uptrace/bun"
)

func (p paymentApp) handleTransactionRequest(ctx context.Context, event entity.Event) error {
	cmd := command.PaymentUserTransactionRequestCommand{}
	if err := json.Unmarshal(event.Payload, &cmd); err != nil {
		return fmt.Errorf("app.payment.handleTransactionRequest: failed to unmarshal command: %w", err)
	}
	ctx = utils.SetRequestTime(ctx, event.CreateTime)

	var userPayment entity.UserPayment
	var transactionRequest entity.UserPaymentTransactionRequest

	err := p.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		tr, err := p.repository.payment.GetPaymentTransactionRequest(ctx, i, cmd.OrderId)
		if err != nil && !errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("app.payment.handleTransactionRequest: failed to get transaction request: %w", err)
		}
		if tr.OrderId != "" {
			return nil
		}

		up, err := p.repository.payment.GetUserPayment(ctx, i, cmd.PaymentId)
		if err != nil {
			return fmt.Errorf("app.payment.handleTransactionRequest: error while get user payment: %w", err)
		}

		if up.UserId != cmd.UserId {
			return fmt.Errorf("app.payment.handleTransactionRequest: invalid user payment: %w", value.ErrUnAuthorized)
		}
		userPayment = up

		tr = entity.UserPaymentTransactionRequest{
			OrderId:            cmd.OrderId,
			UserId:             cmd.UserId,
			PaymentSummary:     userPayment.ToSummary(),
			OrderName:          cmd.OrderName,
			Amount:             cmd.Amount,
			UsedPoint:          cmd.UsedPoint,
			SettlementAmount:   cmd.SettlementAmount,
			SettlementTargetId: cmd.SettlementTargetId,
			Recovery:           cmd.Recovery,
			CreateTime:         event.CreateTime,
		}

		if err := p.repository.payment.CreatePaymentTransactionRequest(ctx, i, tr); err != nil {
			return fmt.Errorf("app.payment.handleTransactionRequest: error while create payment request: %w", err)
		}

		transactionRequest = tr

		return nil
	})

	if err != nil {
		return err
	}

	isMockPayment := userPayment.PaymentType == enum.PaymentType_Mock
	isSignupPromotionPayment := userPayment.PaymentType == enum.PaymentType_SignupPromition
	isTrasactionUnderZero := cmd.TransactionAmount() <= 0

	if isMockPayment || isSignupPromotionPayment || isTrasactionUnderZero {
		// orderId, paymentKey, receiptUrl string, createTime time.Time
		successCmd := command.NewUserPaymentTransactionSuccessCommand(
			transactionRequest.OrderId,
			"",
			"",
			event.CreateTime,
		)
		return p.Run(ctx, func(ctx context.Context, i bun.IDB) error {
			if err := p.repository.event.BatchCreate(ctx, i, []entity.Event{successCmd}); err != nil {
				return fmt.Errorf("app.payment.handleTransactionRequest: error while create mock payment success event: %w", err)
			}
			return nil
		})
	}

	paymentTransaction := value.Payment{
		OrderId:   transactionRequest.OrderId,
		Amount:    transactionRequest.GetPaymentAmount(),
		OrderName: transactionRequest.OrderName,
	}

	_, err = p.service.payment.Transaction(ctx, userPayment, paymentTransaction)
	if errors.Is(err, value.ErrPaymentDuplicatedOrder) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("app.payment.handleTransactionRequest: error while execute payment trasaction: %w", err)
	}

	return nil
}

func (p paymentApp) handleTransactionSuccess(ctx context.Context, event entity.Event) error {
	cmd := command.PaymentUserTransactionSuccessCommand{}
	if err := json.Unmarshal(event.Payload, &cmd); err != nil {
		return fmt.Errorf("app.payment.handleTransactionSuccess: failed to unmarshal command: %w", err)
	}

	err := p.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		var events []entity.Event
		transactionRequest, err := p.repository.payment.GetPaymentTransactionRequest(ctx, i, cmd.OrderId)
		if errors.Is(err, value.ErrNotFound) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("app.payment.handleTransactionSuccess: failed to get transaction request: %w", err)
		}

		if err := p.repository.payment.DeletePaymentTransactionRequest(ctx, i, transactionRequest); err != nil {
			return fmt.Errorf("app.payment.handleTransactionSuccess: failed to delete transaction request: %w", err)
		}

		userPaymentOrder := entity.UserPaymentOrder{
			OrderId:        transactionRequest.OrderId,
			UserId:         transactionRequest.UserId,
			PaymentSummary: transactionRequest.PaymentSummary,
			OrderName:      transactionRequest.OrderName,
			Amount:         transactionRequest.Amount,
			UsedPoint:      transactionRequest.UsedPoint,
			PaymentKey:     cmd.PaymentKey,
			ReceiptUrl:     cmd.ReceiptUrl,
			CreateTime:     cmd.CreateTime,
		}

		if err := p.repository.payment.CreatePaymentOrder(ctx, i, userPaymentOrder); err != nil {
			return fmt.Errorf("app.payment.handleTransactionSuccess: failed to create user payment order: %w", err)
		}

		userPayment, err := p.repository.payment.GetUserPayment(ctx, i, userPaymentOrder.PaymentSummary.PaymentId)
		if err != nil {
			return fmt.Errorf("app.payment.handleTransactionSuccess: failed to get user payment: %w", err)
		}

		switch userPayment.PaymentType {
		case enum.PaymentType_Mock:
			if err := p.AddUserPaymentPoint(ctx, transactionRequest.UserId, transactionRequest.UsedPoint); err != nil {
				return fmt.Errorf("app.payment.handleTransactionSuccess: failed to cancel user used point for mock payment: %w", err)
			}

			return nil
		case enum.PaymentType_SignupPromition:
			if err := p.repository.payment.DeleteUserPayment(ctx, i, userPayment.Id); err != nil {
				return fmt.Errorf("app.payment.handleTransactionSuccess: failed to delete user promotion payment: %w", err)
			}
		}

		_, err = p.ApplyUserReferralReward(ctx, transactionRequest.UserId, transactionRequest.OrderId, transactionRequest.Amount)
		if err != nil {
			return fmt.Errorf("app.payment.handleTransactionSuccess: failed to apply user referral: %w", err)
		}

		_, err = p.ApplyDriverReferralReward(ctx, transactionRequest.SettlementTargetId, transactionRequest.OrderId, transactionRequest.Amount)
		if err != nil {
			return fmt.Errorf("app.payment.handleTransactionSuccess: failed to apply driver referral: %w", err)
		}

		if transactionRequest.Amount < 0 {
			userPaymentPoint, err := p.repository.payment.GetUserPaymentPoint(ctx, i, transactionRequest.UserId)
			if err != nil {
				return fmt.Errorf("app.payment.handleTransactionSuccess: failed to get user payment point: %w", err)
			}
			userPaymentPoint.Point -= transactionRequest.Amount
			if err := p.repository.payment.UpdateUserPaymentPoint(ctx, i, userPaymentPoint); err != nil {
				return fmt.Errorf("app.payment.handleTransactionSuccess: failed to add user payment point: %w", err)
			}
			events = append(events, NewUserPaymentPointAddNotification(transactionRequest.UserId, -transactionRequest.Amount))
		}

		if transactionRequest.SettlementTargetId != uuid.Nil.String() {
			events = append(events, command.NewDriverSettlementRequestCommand(
				transactionRequest.SettlementTargetId,
				transactionRequest.OrderId,
				transactionRequest.GetSettlementAmount(),
				cmd.CreateTime,
			))
		}

		if transactionRequest.Recovery {
			err = p.repository.payment.DeleteFailedOrder(ctx, i, entity.UserPaymentFailedOrder{OrderId: transactionRequest.OrderId})
			if err != nil {
				return fmt.Errorf("app.payment.handleTransactionSuccess: failed to delete failed order: %w", err)
			}

			failedOrders, err := p.repository.payment.GetFailedOrdersByUserId(ctx, i, transactionRequest.UserId)
			if err != nil {
				return fmt.Errorf("app.payment.handleTransactionSuccess: failed to list failed order: %w", err)
			}

			if len(failedOrders) == 0 {
				userPayment, err := p.repository.payment.GetUserPayment(ctx, i, transactionRequest.PaymentSummary.PaymentId)
				if err != nil {
					return fmt.Errorf("app.payment.handleTransactionSuccess: failed to get user payment: %w", err)
				}

				userPayment.Invalid = false
				userPayment.InvalidErrorCode = ""
				userPayment.InvalidErrorMessage = ""

				if err := p.repository.payment.UpdateUserPayment(ctx, i, userPayment); err != nil {
					return fmt.Errorf("app.payment.handleTransactionSuccess: failed to update user payment to valid: %w", err)
				}

			} else {
				events = append(events, command.NewUserPaymentTransactionRequestCommand(
					transactionRequest.UserId,
					transactionRequest.PaymentSummary.PaymentId,
					failedOrders[0].OrderId,
					failedOrders[0].OrderName,
					failedOrders[0].SettlementTargetId,
					failedOrders[0].Amount,
					failedOrders[0].UsedPoint,
					failedOrders[0].SettlementAmount,
					true,
				))
			}
		}

		if err := p.repository.event.BatchCreate(ctx, i, events); err != nil {
			return fmt.Errorf("app.payment.handleTransactionSuccess: failed to create consecutive payment recovery command: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (p paymentApp) handleTransactionFail(ctx context.Context, event entity.Event) error {
	cmd := command.PaymentUserTransactionFailCommand{}
	if err := json.Unmarshal(event.Payload, &cmd); err != nil {
		return fmt.Errorf("app.payment.handleTransactionFail: failed to unmarshal transaction command: %w", err)
	}

	err := p.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		transactionRequest, err := p.repository.payment.GetPaymentTransactionRequest(ctx, i, cmd.OrderId)
		if err != nil {
			return fmt.Errorf("app.payment.handleTransactionFail: failed to get transaction request: %w", err)
		}

		if err := p.repository.payment.DeletePaymentTransactionRequest(ctx, i, transactionRequest); err != nil {
			return fmt.Errorf("app.payment.handleTransactionFail: failed to delete transaction request: %w", err)
		}

		// Set payment as invalid
		userPayment, err := p.repository.payment.GetUserPayment(ctx, i, transactionRequest.PaymentSummary.PaymentId)
		if err != nil {
			return fmt.Errorf("app.payment.handleTransactionFail: error while get user pamyment: %w", err)
		}

		userPayment.Invalid = true
		userPayment.InvalidErrorCode = cmd.FailureCode
		userPayment.InvalidErrorMessage = cmd.FailureReason

		if err := p.repository.payment.UpdateUserPayment(ctx, i, userPayment); err != nil {
			return fmt.Errorf("app.payment.handleTransactionFail: error while update user payment: %w", err)
		}

		// If there is another payment, use it first
		userPayments, err := p.repository.payment.ListUserPayment(ctx, i, transactionRequest.UserId)
		if err != nil {
			return fmt.Errorf("app.payment.handleTransactionFail: error while list user payment: %w", err)
		}

		for _, fallbackUserPayment := range userPayments {
			if fallbackUserPayment.Id != userPayment.Id && !fallbackUserPayment.Invalid {
				fallbackTransactionRequest := command.NewUserPaymentTransactionRequestCommand(
					transactionRequest.UserId,
					fallbackUserPayment.Id,
					transactionRequest.OrderId,
					transactionRequest.OrderName,
					transactionRequest.SettlementTargetId,
					transactionRequest.Amount,
					transactionRequest.UsedPoint,
					transactionRequest.SettlementAmount,
					false,
				)
				fallbackTransactionPush := NewPaymentFallbackNotification(userPayment, fallbackUserPayment)
				if err := p.repository.event.BatchCreate(ctx, i, []entity.Event{fallbackTransactionRequest, fallbackTransactionPush}); err != nil {
					return fmt.Errorf("app.payment.handleTransactionFail: error while create fallback events: %w", err)
				}
				return nil
			}
		}

		// There is no valid payment
		failedOrder := entity.UserPaymentFailedOrder{
			OrderId:            transactionRequest.OrderId,
			UserId:             transactionRequest.UserId,
			OrderName:          transactionRequest.OrderName,
			SettlementTargetId: transactionRequest.SettlementTargetId,
			Amount:             transactionRequest.Amount,
			UsedPoint:          transactionRequest.UsedPoint,
			SettlementAmount:   transactionRequest.SettlementAmount,
			CreateTime:         event.CreateTime,
		}

		if err := p.repository.payment.CreateFailedOrder(ctx, i, failedOrder); err != nil {
			return fmt.Errorf("app.payment.handleTransactionFail: error while create failed order: %w", err)
		}

		failedNotification := NewPaymentFailedNotification(transactionRequest.UserId)
		if err := p.repository.event.BatchCreate(ctx, i, []entity.Event{failedNotification}); err != nil {
			return fmt.Errorf("app.payment.handleTransactionFail: error while create failure event: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (p paymentApp) handleDeletePayment(ctx context.Context, ev entity.Event) error {
	cmd := command.PaymentUserPaymentDeleteCommand{}
	if err := json.Unmarshal(ev.Payload, &cmd); err != nil {
		return fmt.Errorf("app.payment.handleDeletePayment: failed to unmarshal command: %w", err)
	}

	if err := p.service.payment.DeleteCard(ctx, cmd.BillingKey); err != nil {
		return fmt.Errorf("app.payment.handleDeletePayment: failed to delete from payment service: %w", err)
	}

	return nil
}
