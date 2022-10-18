package push

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/event/command"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/domain/value/enum"
	"github.com/uptrace/bun"
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
				fmt.Printf("[TaxiCallPushApp.Worker] error while consume event: %+v\n", err)
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

	// if event.RetryCount > 2 {
	// 	// TODO (taekyeom) attempt count be paramterized
	// 	fmt.Println("[TaxiCallPushApp.Worker] ignore event attempt count limit reached")
	// 	event.Ack()
	// 	return nil
	// }

	defer event.Ack()

	switch event.EventUri {
	case command.EventUri_UserTaxiCallNotification:
		err = t.handleUserNotification(ctx, event)
	case command.EventUri_DriverTaxiCallNotification:
		err = t.handleDriverNotification(ctx, event)
	}

	// If error occurred, resend event with increased retry event count
	if err != nil && event.RetryCount < 3 {
		newEvent := event.NewEventWithRetry()
		t.service.eventPub.SendMessage(ctx, newEvent)
		return err
	}
	return nil
}

func (t taxiCallPushApp) handleUserNotification(ctx context.Context, event entity.Event) error {
	userNotificationCommand := command.UserTaxiCallNotificationCommand{}
	err := json.Unmarshal(event.Payload, &userNotificationCommand)
	if err != nil {
		return fmt.Errorf("app.taxiCallPushApp.handleUserNotification: erorr while unmarshal user notificaiton event: %w, %v", value.ErrInternal, err)
	}

	var fcmToken string
	err = t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		token, err := t.repository.pushToken.Get(ctx, i, userNotificationCommand.UserId)
		if err != nil {
			return fmt.Errorf("app.taxiCallPushApp.handleUserNotification: error while get fcm token: %w", err)
		}
		fcmToken = token.FcmToken
		return nil
	})

	var notification value.Notification
	switch enum.FromTaxiCallStateString(userNotificationCommand.TaxiCallState) {
	case enum.TaxiCallState_Requested:
		notification, err = t.handleUserTaxiCallRequestProgress(ctx, fcmToken, event.CreateTime, userNotificationCommand)
	case enum.TaxiCallState_DRIVER_TO_DEPARTURE:
		notification, err = t.handleUserTaxiCallRequestAccepted(ctx, fcmToken, event.CreateTime, userNotificationCommand)
	case enum.TaxiCallState_FAILED:
		notification, err = t.handleUserTaxiCallRequestFailed(ctx, fcmToken, event.CreateTime, userNotificationCommand)
	case enum.TaxiCallState_DONE:
		notification, err = t.handleUserTaxiCallRequestDone(ctx, fcmToken, event.CreateTime, userNotificationCommand)
	default:
		return fmt.Errorf("app.taxiCallPushApp.handleUserNotification: unsupported event: %s: %w", userNotificationCommand.TaxiCallState, value.ErrInvalidOperation)
	}

	if err != nil {
		return fmt.Errorf("app.taxiCallPushApp.handleUserNotification: error while handle command: %w", err)
	}

	if err := t.service.notification.SendNotification(ctx, notification); err != nil {
		return fmt.Errorf("app.taxiCallPushApp.handleUserNotification: error while send notification: %w", err)
	}

	return nil
}

func (t taxiCallPushApp) handleDriverNotification(ctx context.Context, event entity.Event) error {
	driverNotificationCommand := command.DriverTaxiCallNotificationCommand{}
	err := json.Unmarshal(event.Payload, &driverNotificationCommand)
	if err != nil {
		return fmt.Errorf("app.taxiCallPushApp.handleDriverNotification: erorr while unmarshal driver notificaiton event: %w, %v", value.ErrInternal, err)
	}

	var fcmToken string
	err = t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		token, err := t.repository.pushToken.Get(ctx, i, driverNotificationCommand.DriverId)
		if err != nil {
			return fmt.Errorf("app.taxiCallPushApp.handleUserNotification: error while get fcm token: %w", err)
		}
		fcmToken = token.FcmToken
		return nil
	})

	var notification value.Notification
	switch enum.FromTaxiCallStateString(driverNotificationCommand.TaxiCallState) {
	case enum.TaxiCallState_Requested:
		notification, err = t.handleDriverTaxiCallRequestTicketDistribution(ctx, fcmToken, event.CreateTime, driverNotificationCommand)
	default:
		return fmt.Errorf("app.taxiCallPushApp.handleDriverNotification: unsupported event: %s: %w", driverNotificationCommand.TaxiCallState, value.ErrInvalidOperation)
	}

	if err != nil {
		return fmt.Errorf("app.taxiCallPushApp.handleDriverNotification: error while handle command: %w", err)
	}

	if err := t.service.notification.SendNotification(ctx, notification); err != nil {
		return fmt.Errorf("app.taxiCallPushApp.handleDriverNotification: error while send notification: %w", err)
	}

	return nil
}
