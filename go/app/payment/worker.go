package payment

import (
	"context"
	"fmt"
	"strings"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/event/command"
	"github.com/taco-labs/taco/go/domain/value"
)

func (p paymentApp) Accept(ctx context.Context, event entity.Event) bool {
	return strings.HasPrefix(event.EventUri, command.EventUri_PaymentPrefix)
}

func (p paymentApp) OnFailure(ctx context.Context, event entity.Event, lastErr error) error {
	return nil
}

func (p paymentApp) Process(ctx context.Context, event entity.Event) error {
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
