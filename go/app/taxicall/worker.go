package taxicall

import (
	"context"
	"strings"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/event/command"
)

func (t taxicallApp) Accept(ctx context.Context, event entity.Event) bool {
	return strings.HasPrefix(event.EventUri, command.EventUri_TaxiCallPrefix)
}

func (t taxicallApp) OnFailure(ctx context.Context, event entity.Event, lastErr error) error {
	return nil
}

func (t taxicallApp) Process(ctx context.Context, event entity.Event) error {
	select {
	case <-ctx.Done():
		return nil
	default:
		return t.handleEvent(ctx, event)
	}
}
