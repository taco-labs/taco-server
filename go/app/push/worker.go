package push

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/event/command"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/domain/value/enum"
	"github.com/uptrace/bun"
)

func (t taxiCallPushApp) Accept(ctx context.Context, event entity.Event) bool {
	return strings.HasPrefix(event.EventUri, command.EventUri_PushPrefix)
}

func (t taxiCallPushApp) OnFailure(ctx context.Context, event entity.Event, lastErr error) error {
	return nil
}

func (t taxiCallPushApp) Process(ctx context.Context, event entity.Event) error {
	select {
	case <-ctx.Done():
		return nil
	default:
		var err error
		switch event.EventUri {
		case command.EventUri_RawMessage:
			err = t.handleRawMessage(ctx, event)
		case command.EventUri_UserTaxiCallNotification:
			err = t.handleUserNotification(ctx, event)
		case command.EventUri_DriverTaxiCallNotification:
			err = t.handleDriverNotification(ctx, event)
		}
		return err
	}
}

func (t taxiCallPushApp) handleRawMessage(ctx context.Context, event entity.Event) error {
	cmd := command.PushRawCommand{}
	err := json.Unmarshal(event.Payload, &cmd)
	if err != nil {
		return fmt.Errorf("app.taxicallPushApp.handleRawMessage: error while unmarshal notification: %w", err)
	}

	var fcmToken string
	err = t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		token, err := t.repository.pushToken.Get(ctx, i, cmd.AccountId)
		if err != nil {
			return fmt.Errorf("app.taxiCallPushApp.handleRawMessage: error while get push token: %w", err)
		}
		fcmToken = token.FcmToken
		return nil
	})

	if err != nil {
		return err
	}

	notification := value.Notification{
		Principal: fcmToken,
		Category:  cmd.Category,
		Message: value.NotificationMessage{
			Title: cmd.MessageTitle,
			Body:  cmd.MessageBody,
		},
		Data: cmd.Data,
	}

	if err := t.service.notification.SendNotification(ctx, notification); err != nil {
		return fmt.Errorf("app.taxiCallPushApp.handleRawMessage: error while send notification: %w", err)
	}

	return nil
}

func (t taxiCallPushApp) handleUserNotification(ctx context.Context, event entity.Event) error {
	userNotificationCommand := command.PushUserTaxiCallCommand{}
	err := json.Unmarshal(event.Payload, &userNotificationCommand)
	if err != nil {
		return fmt.Errorf("app.taxiCallPushApp.handleUserNotification: erorr while unmarshal user notificaiton event: %w, %v", value.ErrInternal, err)
	}

	var fcmToken string
	err = t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		token, err := t.repository.pushToken.Get(ctx, i, userNotificationCommand.UserId)
		if err != nil {
			return fmt.Errorf("app.taxiCallPushApp.handleUserNotification: error while get push token: %w", err)
		}
		fcmToken = token.FcmToken
		return nil
	})

	if err != nil {
		return err
	}

	var notification value.Notification
	switch enum.FromTaxiCallStateString(userNotificationCommand.TaxiCallState) {
	case enum.TaxiCallState_Requested:
		notification, err = t.handleUserTaxiCallRequestProgress(ctx, fcmToken, event.CreateTime, userNotificationCommand)
	case enum.TaxiCallState_DRIVER_TO_DEPARTURE:
		notification, err = t.handleUserTaxiCallRequestAccepted(ctx, fcmToken, event.CreateTime, userNotificationCommand)
	case enum.TaxiCallState_DRIVER_TO_ARRIVAL:
		notification, err = t.handleUserTaxiCallDriverToArrival(ctx, fcmToken, event.CreateTime, userNotificationCommand)
	case enum.TaxiCallState_FAILED:
		notification, err = t.handleUserTaxiCallRequestFailed(ctx, fcmToken, event.CreateTime, userNotificationCommand)
	case enum.TaxiCallState_DONE:
		notification, err = t.handleUserTaxiCallRequestDone(ctx, fcmToken, event.CreateTime, userNotificationCommand)
	case enum.TaxiCallState_DRIVER_CANCELLED:
		notification, err = t.handleDriverTaxiCallRequestCanceled(ctx, fcmToken, event.CreateTime, userNotificationCommand)
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
	driverNotificationCommand := command.PushDriverTaxiCallCommand{}
	err := json.Unmarshal(event.Payload, &driverNotificationCommand)
	if err != nil {
		return fmt.Errorf("app.taxiCallPushApp.handleDriverNotification: erorr while unmarshal driver notificaiton event: %w, %v", value.ErrInternal, err)
	}

	var fcmToken string
	err = t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		token, err := t.repository.pushToken.Get(ctx, i, driverNotificationCommand.DriverId)
		if err != nil {
			return fmt.Errorf("app.taxiCallPushApp.handleDriverNotification: error while get push token: %w", err)
		}
		fcmToken = token.FcmToken
		return nil
	})
	if err != nil {
		return err
	}

	var notification value.Notification
	switch enum.FromTaxiCallStateString(driverNotificationCommand.TaxiCallState) {
	case enum.TaxiCallState_Requested:
		notification, err = t.handleDriverTaxiCallRequestTicketDistribution(ctx, fcmToken, event.CreateTime, driverNotificationCommand)
	case enum.TaxiCallState_USER_CANCELLED:
		notification, err = t.handleUserTaxiCallRequestCanceled(ctx, fcmToken, event.CreateTime, driverNotificationCommand)
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
