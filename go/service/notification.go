package service

import "context"

type NotificationService interface {
	SendNotification(context.Context, string, map[string]interface{}) error
}
