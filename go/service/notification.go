package service

import (
	"context"
	firebase "firebase.google.com/go"
)

type NotificationService interface {
	SendNotification(context.Context, string, map[string]interface{}) error
}

type firebaseNotificationService struct {
	app *firebase.App
}

func NewFirebaseNotificationService(firebaseApp *firebase.App) firebaseNotificationService {
	return firebaseNotificationService{
		app: firebaseApp,
	}
}
