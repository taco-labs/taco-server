package payment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/event/command"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/service"
	"github.com/uptrace/bun"
)

func (p paymentApp) Start(ctx context.Context) error {
	go p.loop(ctx)
	return nil
}

func (p paymentApp) Shutdown(ctx context.Context) error {
	<-p.waitCh
	return nil
}

func (p paymentApp) loop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("shutting down [Payment App Consumer] stream...")
			p.waitCh <- struct{}{}
			return
		default:
			err := p.consume(ctx)
			if err != nil {
				fmt.Printf("[PaymentApp.Worker] error while consume event: %+v\n", err)
			}
		}
	}
}

func (p paymentApp) consume(ctx context.Context) error {
	event, err := p.service.eventSub.GetMessage(ctx)
	if err != nil {
		return nil
	}

	return p.service.workerPool.Submit(func() {
		var err error
		var events []entity.Event

		defer func() {
			if errors.Is(err, context.Canceled) {
				return
			}

			for _, ev := range events {
				if publishError := p.service.eventPub.SendMessage(ctx, ev); publishError != nil {
					err = fmt.Errorf("app.payment.consume: error while publish message: %v: %w", publishError, err)
					break
				}
			}

			if err != nil {
				// TODO (taekyeom) error logging
				fmt.Printf("[PaymentApp.Worker.Consume] error while handle consumed message (attempt: %d): %+v\n", event.Attempt, err)
				if event.Attempt < 3 {
					event.Nack()
					return
				}
				// TODO (taekyeom) retry limit logging
			}
			event.Ack()
		}()

		switch event.EventUri {
		case command.EventUri_UserTransaction:
			events, err = p.handleTransaction(ctx, event)
		case command.EventUri_UserTransactionFailed:
			events, err = p.handleFailedTransaction(ctx, event)
		case command.EventUri_UserTransactionRecovery:
			events, err = p.handleRecovery(ctx, event)
		default:
			// TODO(taekyeom) logging
			err = fmt.Errorf("%w: [PaymentApp.Worker.Consume] Invalid EventUri '%v'", value.ErrInvalidOperation, event.EventUri)
		}
	})
}

func (p paymentApp) handleTransaction(ctx context.Context, ev entity.Event) (events []entity.Event, err error) {
	cmd := command.PaymentUserTransactionCommand{}
	if err := json.Unmarshal(ev.Payload, &cmd); err != nil {
		return []entity.Event{}, fmt.Errorf("app.payment.handleTransaction: failed to unmarshal transaction command: %w", err)
	}

	defer func() {
		var tacoErr value.TacoError
		if service.AsPaymentError(err, &tacoErr) && ev.Attempt == 3 {
			failedTransactionEvent := command.NewPaymentUserTransactionFailedCommand(
				cmd.UserId, cmd.PaymentId, cmd.OrderId, cmd.OrderName, cmd.Amount,
				tacoErr.ErrCode, tacoErr.Message,
			)
			// TODO (taekyeom) Push notification
			events = append(events, failedTransactionEvent)
		}
	}()

	var userPayment entity.UserPayment
	var userPaymentOrder entity.UserPaymentOrder

	err = p.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		// Precondition: payment in progress (see order entity, check order is done)
		userPaymentOrder, err = p.repository.payment.GetPaymentOrder(ctx, i, cmd.OrderId)
		if err != nil && !errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("app.payment.handleTransaction: error while get payment order: %w", err)
		}
		if userPaymentOrder.OrderId == cmd.OrderId {
			// TODO (taekyeom) duplication order logging
			return nil
		}

		up, err := p.repository.payment.GetUserPayment(ctx, i, cmd.PaymentId)
		if err != nil {
			return fmt.Errorf("app.payment.handleTransaction: error while get user payment: %w", err)
		}

		if up.UserId != cmd.UserId {
			return fmt.Errorf("app.payment.handleTransaction: invalid user payment: %w", value.ErrUnAuthorized)
		}
		if up.Invalid {
			return fmt.Errorf("app.payment.handleTransaction: invalid user payment: %v, %v %w",
				up.InvalidErrorCode, up.InvalidErrorMessage, value.ErrInvalidUserPayment)
		}
		userPayment = up

		return nil
	})

	if err != nil {
		return []entity.Event{}, err
	}

	paymentTransaction := value.Payment{
		OrderId:   cmd.OrderId,
		Amount:    cmd.Amount,
		OrderName: cmd.OrderName,
	}

	transactionResult, err := p.service.payment.Transaction(ctx, userPayment, paymentTransaction)
	if errors.Is(err, value.ErrPaymentDuplicatedOrder) && userPaymentOrder.OrderId != "" {
		// TODO (taekyeom) duplication order logging
		return []entity.Event{}, nil
	}
	if err != nil && !errors.Is(err, value.ErrPaymentDuplicatedOrder) {
		return []entity.Event{}, fmt.Errorf("app.payment.handleTransaction: error while execute payment trasaction: %w", err)
	}

	if transactionResult.OrderId == "" {
		transactionResult, err = p.service.payment.GetTransactionResult(ctx, cmd.OrderId)
		if err != nil {
			return []entity.Event{}, fmt.Errorf("app.payment.handleTransaction: error while get transaction result from payment servie: %w", err)
		}
	}

	err = p.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		userPaymentOrder := entity.UserPaymentOrder{
			OrderId:        cmd.OrderId,
			UserId:         cmd.UserId,
			PaymentSummary: userPayment.ToSummary(),
			OrderName:      cmd.OrderName,
			Amount:         cmd.Amount,
			PaymentKey:     transactionResult.PaymentKey,
			ReceiptUrl:     transactionResult.ReceiptUrl,
			CreateTime:     ev.CreateTime,
		}

		if err := p.repository.payment.CreatePaymentOrder(ctx, i, userPaymentOrder); err != nil {
			return fmt.Errorf("app.payment.handleTransaction: error while create payment order: %w", err)
		}

		events = append(events, NewPaymentSuccessNotification(userPaymentOrder))

		return nil
	})

	if err != nil {
		return []entity.Event{}, err
	}

	return []entity.Event{}, nil
}

func (p paymentApp) handleFailedTransaction(ctx context.Context, ev entity.Event) ([]entity.Event, error) {
	cmd := command.PaymentUserTransactionFailedCommand{}
	if err := json.Unmarshal(ev.Payload, &cmd); err != nil {
		return []entity.Event{}, fmt.Errorf("app.payment.handleFailedTransaction: failed to unmarshal transaction command: %w", err)
	}

	var events []entity.Event

	err := p.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		// Set payment as invalid
		userPayment, err := p.repository.payment.GetUserPayment(ctx, i, cmd.PaymentId)
		if err != nil {
			return fmt.Errorf("app.payment.handleFailedTransaction: error while get user pamyment: %w", err)
		}

		if userPayment.UserId != cmd.UserId {
			return fmt.Errorf("app.payment.handleFailedTransaction: invalid user payment: %w", value.ErrUnAuthorized)
		}

		userPayment.Invalid = true
		userPayment.InvalidErrorCode = cmd.FailedErrorCode
		userPayment.InvalidErrorMessage = cmd.FailedErrorMessage

		if err := p.repository.payment.UpdateUserPayment(ctx, i, userPayment); err != nil {
			return fmt.Errorf("app.payment.handleFailedTransaction: error while update user payment: %w", err)
		}

		// If there is another payment, use it first
		userPayments, err := p.repository.payment.ListUserPayment(ctx, i, cmd.UserId)
		if err != nil {
			return fmt.Errorf("app.payment.handleFailedTransaction: error while list user payment: %w", err)
		}

		for _, fallbackUserPayment := range userPayments {
			if fallbackUserPayment.Id != cmd.PaymentId && !fallbackUserPayment.Invalid {
				fallbackTransactionCmd := command.NewPaymentUserTransactionCommand(
					cmd.UserId, fallbackUserPayment.Id, cmd.OrderId, cmd.OrderName, cmd.Amount,
				)
				fallbackTransactionPush := NewPaymentFallbackNotification(userPayment, fallbackUserPayment)
				events = append(events, fallbackTransactionCmd, fallbackTransactionPush)
				return nil
			}
		}

		// There is no valid payment...
		failedOrder := entity.UserPaymentFailedOrder{
			OrderId:    cmd.OrderId,
			UserId:     cmd.UserId,
			OrderName:  cmd.OrderName,
			Amount:     cmd.Amount,
			CreateTime: ev.CreateTime,
		}

		if err := p.repository.payment.CreateFailedOrder(ctx, i, failedOrder); err != nil {
			return fmt.Errorf("app.payment.handleFailedTransaction: error while create failed order: %w", err)
		}

		events = append(events, NewPaymentFailedNotification(cmd.UserId))

		return nil
	})

	return events, err
}

func (p paymentApp) handleRecovery(ctx context.Context, ev entity.Event) ([]entity.Event, error) {
	cmd := command.PaymentUserTransactionRecoveryCommand{}
	if err := json.Unmarshal(ev.Payload, &cmd); err != nil {
		return []entity.Event{}, fmt.Errorf("app.payment.handleRecovery: failed to unmarshal transaction command: %w", err)
	}

	var events []entity.Event

	// Set payment as invalid
	err := p.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		userPayment, err := p.repository.payment.GetUserPayment(ctx, i, cmd.PaymentId)
		if err != nil {
			return fmt.Errorf("app.payment.handleRecovery: error while get user pamyment: %w", err)
		}

		if userPayment.UserId != cmd.UserId {
			return fmt.Errorf("app.payment.handleRecovery: invalid user payment: %w", value.ErrUnAuthorized)
		}

		failedOrders, err := p.repository.payment.GetFailedOrdersByUserId(ctx, i, cmd.UserId)
		if err != nil {
			return fmt.Errorf("app.payment.handleRecovery: error while get failed orders: %w", err)
		}

		// TODO (taekyeom) optimize it...
		for _, failedOrder := range failedOrders {
			userPaymentOrder, err := p.repository.payment.GetPaymentOrder(ctx, i, failedOrder.OrderId)
			if err != nil && !errors.Is(err, value.ErrNotFound) {
				return fmt.Errorf("app.payment.handleRecovery: error while get payment order: %w", err)
			}
			if userPaymentOrder.OrderId == failedOrder.OrderId {
				continue
			}
			paymentTransaction := value.Payment{
				// OrderId   string
				// Amount    int
				// OrderName string
				OrderId:   failedOrder.OrderId,
				Amount:    failedOrder.Amount,
				OrderName: failedOrder.OrderName,
			}
			transactionResult, err := p.service.payment.Transaction(ctx, userPayment, paymentTransaction)
			if errors.Is(err, value.ErrPaymentDuplicatedOrder) && userPaymentOrder.OrderId != "" {
				// TODO (taekyeom) duplication order logging
				continue
			}
			if err != nil && !errors.Is(err, value.ErrPaymentDuplicatedOrder) {
				return fmt.Errorf("app.payment.handleRecovery: error while execute payment trasaction: %w", err)
			}

			if transactionResult.OrderId == "" {
				transactionResult, err = p.service.payment.GetTransactionResult(ctx, failedOrder.OrderId)
				if err != nil {
					return fmt.Errorf("app.payment.handleRecovery: error while get transaction result from payment servie: %w", err)
				}
			}

			userPaymentOrder = entity.UserPaymentOrder{
				OrderId:        failedOrder.OrderId,
				UserId:         cmd.UserId,
				PaymentSummary: userPayment.ToSummary(),
				OrderName:      failedOrder.OrderName,
				Amount:         failedOrder.Amount,
				PaymentKey:     transactionResult.PaymentKey,
				ReceiptUrl:     transactionResult.ReceiptUrl,
				CreateTime:     ev.CreateTime,
			}

			if err := p.repository.payment.CreatePaymentOrder(ctx, i, userPaymentOrder); err != nil {
				return fmt.Errorf("app.payment.handleRecovery: error while create payment order: %w", err)
			}

			if err := p.repository.payment.DeleteFailedOrder(ctx, i, failedOrder); err != nil {
				return fmt.Errorf("app.payment.handleRecovery: error while delete failed order: %w", err)
			}
		}

		userPayment.Invalid = false
		userPayment.InvalidErrorCode = ""
		userPayment.InvalidErrorMessage = ""
		if err := p.repository.payment.UpdateUserPayment(ctx, i, userPayment); err != nil {
			return fmt.Errorf("app.payment.handleRecovery: error while update user payment: %w", err)
		}

		events = append(events, NewPaymentRecoveryNotification(userPayment))

		return nil
	})

	return []entity.Event{}, err
}
