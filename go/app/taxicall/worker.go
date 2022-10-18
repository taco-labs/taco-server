package taxicall

import (
	"context"
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
				fmt.Printf("[TaxiCallApp.Worker] error while consume event: %+v", err)
			}
		}
	}
}

func (t taxicallApp) consume(ctx context.Context) error {
	event, err := t.service.eventSub.GetMessage(ctx)
	if err != nil {
		return nil
	}
	defer event.Ack()

	err = t.handleEvent(ctx, event)
	if err != nil && event.RetryCount < 3 {
		newEvent := event.NewEventWithRetry()
		newEvent.DelaySeconds = 0
		t.service.eventPub.SendMessage(ctx, newEvent)
		return err
	}

	return nil
}
