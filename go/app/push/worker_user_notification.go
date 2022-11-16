package push

import (
	"context"
	"fmt"
	"time"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/event/command"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/uptrace/bun"
)

func (t taxiCallPushApp) handleUserTaxiCallRequestProgress(ctx context.Context, fcmToken string,
	eventTime time.Time, cmd command.PushUserTaxiCallCommand) (value.Notification, error) {
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
	eventTime time.Time, cmd command.PushUserTaxiCallCommand) (value.Notification, error) {
	// Get Driver
	var driver entity.Driver
	err := t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		dr, err := t.service.driverGetter.GetDriver(ctx, cmd.DriverId)
		if err != nil {
			return fmt.Errorf("app.push.handleUserTaxiCallRequestAccepted: error while get driver: %w", err)
		}
		driver = dr
		return nil
	})
	if err != nil {
		return value.Notification{}, err
	}

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
		"driverId":              cmd.DriverId,
		"driverPhone":           driver.Phone,
		"driverCarNumber":       driver.CarNumber,
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
	eventTime time.Time, cmd command.PushUserTaxiCallCommand) (value.Notification, error) {

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
	eventTime time.Time, cmd command.PushUserTaxiCallCommand) (value.Notification, error) {

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

func (t taxiCallPushApp) handleDriverTaxiCallRequestCanceled(ctx context.Context, fcmToken string,
	eventTime time.Time, cmd command.PushUserTaxiCallCommand) (value.Notification, error) {
	message := value.NotificationMessage{
		Title: "운행 취소",
		Body:  "기사님이 택시 운행을 취소하였습니다.",
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
