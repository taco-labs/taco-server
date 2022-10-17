package push

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/event/command"
	"github.com/taco-labs/taco/go/domain/value"
)

func (t taxiCallPushApp) Start(ctx context.Context) error {
	go t.loop(ctx)
	return nil
}

func (t taxiCallPushApp) Stop(ctx context.Context) error {
	<-t.waitCh
	return nil
}

func (t taxiCallPushApp) loop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("shutting down [Taxi Call Push Consumer] stream...")
			t.waitCh <- struct{}{}
			return
		default:
			err := t.consume(ctx)
			if err != nil {
				//TODO (taekyeom) logging
			}
		}
	}
}

// TODO (taekyeom) parallelize
func (t taxiCallPushApp) consume(ctx context.Context) error {
	event, err := t.service.eventSub.GetMessage(ctx)
	if err != nil {
		return nil
	}
	fmt.Printf("Consume:: %++v\n", event)
	if event.AttemtCount > 3 {
		// TODO (taekyeom) attempt count be paramterized
		event.Ack()
		return nil
	}
	switch event.EventUri {
	case command.EventUri_UserTaxiCallNotification:
		err = t.handleUserNotification(ctx, event)
	case command.EventUri_DriverTaxiCallNotification:
		err = t.handleDriverNotification(ctx, event)
	}
	if err != nil {
		return err
	}
	return event.Ack()
}

func (t taxiCallPushApp) handleUserNotification(ctx context.Context, event entity.Event) error {
	userNotificationCommand := command.UserTaxiCallNotificationCommand{}
	err := json.Unmarshal(event.Payload, &userNotificationCommand)
	if err != nil {
		return fmt.Errorf("app.taxiCallPushApp.handleUserNotification: erorr while unmarshal user notificaiton event: %w, %v", value.ErrInternal, err)
	}
	return nil
}

func (t taxiCallPushApp) handleDriverNotification(ctx context.Context, event entity.Event) error {
	return nil
}
