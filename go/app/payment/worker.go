package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/event/command"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/uptrace/bun"
)

func (p paymentApp) Accept(ctx context.Context, event entity.Event) bool {
	return strings.HasPrefix(event.EventUri, command.EventUri_PaymentPrefix)
}

func (p paymentApp) OnFailure(ctx context.Context, event entity.Event, lastErr error) error {
	var err error
	switch event.EventUri {
	case command.EventUri_UserTransactionRequest:
		cmd := command.PaymentUserTransactionRequestCommand{}
		err = json.Unmarshal(event.Payload, &cmd)
		if err != nil {
			return fmt.Errorf("app.taxicall.OnFailure: error while unmarshal json: %v", err)
		}
		err = p.makeTransactionFail(ctx, cmd.OrderId, "결제 요청에 실패했습니다")
	}
	// TODO (taekyeom) transaction success 시 어떻게 처리해야 할까..? (일단은 수동으로 처리하자)
	return err
}

func (p paymentApp) makeTransactionFail(ctx context.Context, orderId string, failureReason string) error {
	return p.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		cmd := command.NewUserPaymentTransactionFailCommand(orderId, string(value.ERR_EXTERNAL), failureReason)

		if err := p.repository.event.BatchCreate(ctx, i, []entity.Event{cmd}); err != nil {
			return fmt.Errorf("app.payment.makeTransactionFail: error while make transaction fail: %w", err)
		}

		return nil
	})
}

func (p paymentApp) Process(ctx context.Context, event entity.Event) error {
	requestTime := time.Now()
	defer func() {
		tags := []string{"eventUri", event.EventUri}
		now := time.Now()
		p.service.metric.Timing("WorkerProcessTime", now.Sub(requestTime), tags...)
	}()

	select {
	case <-ctx.Done():
		return nil
	default:
		return p.handleEvent(ctx, event)
	}
}

func (p paymentApp) handleEvent(ctx context.Context, event entity.Event) error {
	var err error

	switch event.EventUri {
	case command.EventUri_UserTransactionRequest:
		err = p.handleTransactionRequest(ctx, event)
	case command.EventUri_UserTransactionSuccess:
		err = p.handleTransactionSuccess(ctx, event)
	case command.EventUri_UserTransactionFail:
		err = p.handleTransactionFail(ctx, event)
	case command.EventUri_UserDeletePayment:
		err = p.handleDeletePayment(ctx, event)
	default:
		err = fmt.Errorf("%w: [PaymentApp.Worker.Consume] Invalid EventUri '%v'", value.ErrInvalidOperation, event.EventUri)
	}

	return err
}
