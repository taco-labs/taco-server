package payment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/event/command"
	"github.com/taco-labs/taco/go/domain/value"
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
		switch event.EventUri {
		case command.EventUri_UserTransaction:
			if event.Attempt < 4 {
				err = p.handleTransaction(ctx, event)
			} else {
				// TODO (taekyeom) handle failed transaction이 실패하면?
				err = p.handleFailedTransaction(ctx, event)
			}
		default:
			// TODO(taekyeom) logging
			fmt.Printf("[PaymentApp.Worker.Consume] Invalid EventUri '%v'", event.EventUri)
		}
		if errors.Is(err, context.Canceled) {
			return
		}
		if err != nil {
			fmt.Printf("[PaymentApp.Worker.Consume] error while handle consumed message: %+v\n", err)
			event.Nack()
			return
		}
		event.Ack()
	})
}

func (p paymentApp) handleTransaction(ctx context.Context, ev entity.Event) error {
	cmd := command.PaymentUserTransactionCommand{}
	if err := json.Unmarshal(ev.Payload, &cmd); err != nil {
		return fmt.Errorf("app.payment.handleTransaction: failed to unmarshal transaction command: %w", err)
	}

	var userPayment entity.UserPayment
	err := p.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		up, err := p.repository.payment.GetUserPayment(ctx, i, cmd.PaymentId)
		if err != nil {
			return fmt.Errorf("app.payment.handleTransaction: error while get user payment: %w", err)
		}

		if up.UserId != cmd.UserId {
			return fmt.Errorf("app.payment.handleTransaction: invalid user payment: %w", value.ErrUnAuthorized)
		}
		userPayment = up

		return nil
	})

	if err != nil {
		return err
	}

	fmt.Printf("HandleTransaction::%+v\n", cmd)

	paymentTransaction := value.Payment{
		OrderId:   cmd.OrderId,
		Amount:    cmd.Amount,
		OrderName: cmd.OrderName,
	}

	// TODO (taekyeom) order id 중복 요청 핸들링
	_, err = p.service.payment.Transaction(ctx, userPayment, paymentTransaction)
	if err != nil {
		return fmt.Errorf("app.payment.handleTransaction: error while execute payment trasaction: %w", err)
	}

	// TODO (taekyeom) payment result handling
	// p.Run(ctx, func(ctx context.Context, i bun.IDB) error {
	// })

	return nil
}

func (p paymentApp) handleFailedTransaction(ctx context.Context, ev entity.Event) error {
	return nil
}
