package service

import (
	"context"
	"fmt"

	"firebase.google.com/go/messaging"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/utils/slices"
)

type NotificationService interface {
	SendNotification(context.Context, string, map[string]string) error
}

type firebaseNotificationService struct {
	client *messaging.Client
	dryRun bool
}

func (f firebaseNotificationService) SendNotification(ctx context.Context, notification value.Notification) error {
	fcmMessage := notificationToFcmMessage(notification)
	// TODO (taekyeom) handle return string?
	var err error
	if f.dryRun {
		_, err = f.client.SendDryRun(ctx, fcmMessage)
	} else {
		_, err = f.client.Send(ctx, fcmMessage)
	}
	if err != nil {
		return fmt.Errorf("%w: error while send cloud messaing notification: %v", value.ErrExternal, err)
	}
	return nil
}

func (f firebaseNotificationService) BulkSendNotification(ctx context.Context, notifications []value.Notification) error {
	fcmMessages := slices.Map(notifications, notificationToFcmMessage)
	var err error
	if f.dryRun {
		_, err = f.client.SendAllDryRun(ctx, fcmMessages)
	} else {
		_, err = f.client.SendAll(ctx, fcmMessages)
	}
	if err != nil {
		return fmt.Errorf("%w: error while bulk send cloud messaing notification: %v", value.ErrExternal, err)
	}
	return nil
}

func NewFirebaseNotificationService(client *messaging.Client, dryRun bool) firebaseNotificationService {
	return firebaseNotificationService{
		client: client,
		dryRun: dryRun,
	}
}

func notificationToFcmMessage(notification value.Notification) *messaging.Message {
	message := messaging.Message{
		Token: notification.Principal(),
		Data:  notification.Data(),
	}
	if !notification.DataOnly() {
		notificationMessage := notification.NotificationMessage()
		message.Notification = &messaging.Notification{
			Title: notificationMessage.Title,
			Body:  notificationMessage.Body,
		}
	}

	return &message
}
