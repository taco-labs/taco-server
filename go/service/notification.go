package service

import (
	"context"

	"firebase.google.com/go/messaging"
	"github.com/taco-labs/taco/go/domain/value"
	"gocloud.dev/pubsub"
)

type NotificationService interface {
	SendNotification(context.Context, value.Notification) error
}

type firebaseNotificationService struct {
	pub *pubsub.Topic
}

func (f firebaseNotificationService) SendNotification(ctx context.Context, notification value.Notification) error {
	return f.pub.Send(ctx, notificationToMessage(notification))
}

func NewFirebaseNotificationService(topic *pubsub.Topic) *firebaseNotificationService {
	return &firebaseNotificationService{
		pub: topic,
	}
}

func notificationToMessage(notification value.Notification) *pubsub.Message {
	fcmMessage := notificationToFcmMessage(notification)

	fcmMessagePayload, _ := fcmMessage.MarshalJSON()

	return &pubsub.Message{
		Body: fcmMessagePayload,
	}
}

func notificationToFcmMessage(notification value.Notification) *messaging.Message {
	message := messaging.Message{
		Token: notification.Principal,
		Data:  notification.Data,
	}
	if notification.Message.Title != "" {
		message.Notification = &messaging.Notification{
			Title: notification.Message.Title,
			Body:  notification.Message.Body,
		}
	}

	return &message
}
