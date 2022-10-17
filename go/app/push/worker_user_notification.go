package push

import (
	"context"
	"fmt"
	"time"

	"github.com/taco-labs/taco/go/domain/event/command"
	"github.com/taco-labs/taco/go/domain/value"
)

func (t taxiCallPushApp) handleUserTaxiCallRequestProgress(ctx context.Context, fcmToken string,
	eventTime time.Time, cmd command.UserTaxiCallNotificationCommand) (value.Notification, error) {
	data := map[string]string{
		"taxiCallRequestId":      cmd.TaxiCallRequestId,
		"taxiCallState":          cmd.TaxiCallState,
		"currentAdditionalPrice": fmt.Sprint(cmd.AdditionalPrice),
		"searchRangeInMinutes":   fmt.Sprint(cmd.SearchRangeInMinutes),
	}

	return value.Notification{
		Principal: fcmToken,
		Data:      data,
	}, nil
}

func (t taxiCallPushApp) handleUserTaxiCallRequestAccepted(ctx context.Context, fcmToken string,
	eventTime time.Time, cmd command.UserTaxiCallNotificationCommand) (value.Notification, error) {
	routeBetweenDeparture, err := t.service.route.GetRoute(ctx, cmd.DriverLocation, cmd.Departure.Point)
	if err != nil {
		return value.Notification{},
			fmt.Errorf("app.push.handleUserTaxiCallRequestAccepted: error while get route between driver location and departure: %w", err)
	}

	message := value.NotificationMessage{
		Title: fmt.Sprintf("배차 완료 (약 %d분)", int(routeBetweenDeparture.ETA.Minutes())),
		Body:  fmt.Sprintf("%s (추가 요금 %d)", cmd.Departure.Address.AddressName, cmd.AdditionalPrice),
	}

	data := map[string]string{
		"taxiCallRequestId":     cmd.TaxiCallRequestId,
		"taxiCallState":         cmd.TaxiCallState,
		"driverId":              cmd.DriverId, // TODO (driver informations)
		"requestBasePrice":      fmt.Sprint(cmd.RequestBasePrice),
		"additionalPrice":       fmt.Sprint(cmd.AdditionalPrice),
		"toDepartureDistance":   fmt.Sprint(routeBetweenDeparture.Distance),
		"whenDriverToDeparture": fmt.Sprint(time.Now().Add(routeBetweenDeparture.ETA)),
	}

	return value.Notification{
		Principal: fcmToken,
		Message:   message,
		Data:      data,
	}, nil
}

func (t taxiCallPushApp) handleUserTaxiCallRequestFailed(ctx context.Context, fcmToken string,
	eventTime time.Time, cmd command.UserTaxiCallNotificationCommand) (value.Notification, error) {

	message := value.NotificationMessage{
		Title: "배차 실패",
		Body:  "택시 배차에 실패했습니다",
	}

	data := map[string]string{
		"taxiCallRequestId": cmd.TaxiCallRequestId,
		"taxiCallState":     cmd.TaxiCallState,
	}

	return value.Notification{
		Principal: fcmToken,
		Message:   message,
		Data:      data,
	}, nil
}

func (t taxiCallPushApp) handleUserTaxiCallRequestDone(ctx context.Context, fcmToken string,
	eventTime time.Time, cmd command.UserTaxiCallNotificationCommand) (value.Notification, error) {

	message := value.NotificationMessage{
		Title: "운행 완료",
		Body:  "택시 운행이 완료되었습니다.",
	}

	data := map[string]string{
		"taxiCallRequestId": cmd.TaxiCallRequestId,
		"taxiCallState":     cmd.TaxiCallState,
		"basePrice":         fmt.Sprint(cmd.BasePrice),
		"additionalPrice":   fmt.Sprint(cmd.AdditionalPrice),
	}

	return value.Notification{
		Principal: fcmToken,
		Message:   message,
		Data:      data,
	}, nil
}
