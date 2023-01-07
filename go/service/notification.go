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
	data := notification.Data
	data["category"] = notification.Category

	message := messaging.Message{
		Token: notification.Principal,
		Data:  data,
	}
	message.Android = &messaging.AndroidConfig{}
	message.APNS = &messaging.APNSConfig{}
	message.APNS.Payload = &messaging.APNSPayload{
		Aps: &messaging.Aps{
			ContentAvailable: true,
		},
	}
	if notification.Message.Title != "" {
		message.Notification = &messaging.Notification{
			Title: notification.Message.Title,
			Body:  notification.Message.Body,
		}
	}

	if notification.MessageKey != "" {
		message.Android.CollapseKey = notification.MessageKey
		message.Android.Notification = &messaging.AndroidNotification{
			Tag: notification.MessageKey,
		}

		message.APNS.Headers = map[string]string{
			"apns-collapse-id": notification.MessageKey,
		}
	}

	return &message
}
