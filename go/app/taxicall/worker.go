package taxicall

import (
	"context"
	"errors"
	"fmt"
)

func (t taxicallApp) Start(ctx context.Context) error {
	go t.loop(ctx)
	return nil
}

func (t taxicallApp) Shutdown(ctx context.Context) error {
	<-t.waitCh
	return nil
}

func (t taxicallApp) loop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("shutting down [Taxi Call Consumer] stream...")
			t.waitCh <- struct{}{}
		default:
			err := t.consume(ctx)
			if err != nil {
				//TODO (taekyeom) logging
				fmt.Printf("[TaxicallApp.Worker] error while consume event: %+v\n", err)
			}
		}
	}
}

func (t taxicallApp) consume(ctx context.Context) error {
	event, err := t.service.eventSub.GetMessage(ctx)
	if err != nil {
		return nil
	}
	return t.service.workerPool.Submit(func() {
		err := t.handleEvent(ctx, event)
		if err != nil {
			fmt.Printf("[TaxicallApp.Worker.Consume] error while handle consumed message: %+v\n", err)
			if errors.Is(err, context.Canceled) {
				return
			}
			if event.RetryCount < 3 {
				event.Nack()
			}
		}
		event.Ack()
	})
}
