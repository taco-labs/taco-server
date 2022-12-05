package push

import (
	"context"
	"fmt"
	"time"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/event/command"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/uptrace/bun"
	"golang.org/x/sync/errgroup"
)

func (t taxiCallPushApp) handleUserTaxiCallRequestProgress(ctx context.Context, fcmToken string,
	eventTime time.Time, cmd command.PushUserTaxiCallCommand) (value.Notification, error) {
	data := map[string]string{
		"taxiCallRequestId":      cmd.TaxiCallRequestId,
		"taxiCallState":          cmd.TaxiCallState,
		"currentAdditionalPrice": fmt.Sprint(cmd.AdditionalPrice),
		"searchRangeInMinutes":   fmt.Sprint(cmd.SearchRangeInMinutes),
	}

	return value.NewNotification(fcmToken, value.NotificationCategory_Taxicall, "", "", data), nil
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

	group, gCtx := errgroup.WithContext(ctx)

	var routeBetweenDeparture, routeBetweenArrival value.Route

	group.Go(func() error {
		r, err := t.service.route.GetRoute(gCtx, cmd.DriverLocation, cmd.Departure.Point)
		if err != nil {
			return fmt.Errorf("app.push.handleUserTaxiCallRequestAccepted: error while get route between driver location and departure: %w", err)
		}
		routeBetweenDeparture = r

		return nil
	})

	group.Go(func() error {
		r, err := t.service.route.GetRoute(gCtx, cmd.Departure.Point, cmd.Arrival.Point)
		if err != nil {
			return fmt.Errorf("app.push.handleUserTaxiCallRequestAccepted: error while get route between deprature and arrival: %w", err)
		}
		routeBetweenArrival = r

		return nil
	})

	if err := group.Wait(); err != nil {
		return value.Notification{},
			fmt.Errorf("app.push.handleUserTaxiCallRequestAccepted: error while get route: %w", err)
	}

	messageTitle := fmt.Sprintf("배차 완료 (약 %d분)", int(routeBetweenDeparture.ETA.Minutes()))
	messageBody := fmt.Sprintf("%s (추가 요금 %d)", cmd.Departure.Address.AddressName, cmd.AdditionalPrice)

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
		"toArrivalDistance":     fmt.Sprint(routeBetweenArrival.Distance),
		"toArrivalETA":          fmt.Sprint(routeBetweenArrival.ETA.Nanoseconds()),
	}

	return value.NewNotification(fcmToken, value.NotificationCategory_Taxicall, messageTitle, messageBody, data), nil
}

func (t taxiCallPushApp) handleUserTaxiCallRequestFailed(ctx context.Context, fcmToken string,
	eventTime time.Time, cmd command.PushUserTaxiCallCommand) (value.Notification, error) {

	messageTitle := "배차 실패"
	messageBody := "택시 배차에 실패했습니다"

	data := map[string]string{
		"taxiCallRequestId": cmd.TaxiCallRequestId,
		"taxiCallState":     cmd.TaxiCallState,
	}

	return value.NewNotification(fcmToken, value.NotificationCategory_Taxicall, messageTitle, messageBody, data), nil
}

func (t taxiCallPushApp) handleUserTaxiCallRequestDone(ctx context.Context, fcmToken string,
	eventTime time.Time, cmd command.PushUserTaxiCallCommand) (value.Notification, error) {

	messageTitle := "운행 완료"
	messageBody := "택시 운행이 완료되었습니다."

	data := map[string]string{
		"taxiCallRequestId": cmd.TaxiCallRequestId,
		"taxiCallState":     cmd.TaxiCallState,
		"basePrice":         fmt.Sprint(cmd.BasePrice),
		"additionalPrice":   fmt.Sprint(cmd.AdditionalPrice),
	}

	return value.NewNotification(fcmToken, value.NotificationCategory_Taxicall, messageTitle, messageBody, data), nil
}

func (t taxiCallPushApp) handleDriverTaxiCallRequestCanceled(ctx context.Context, fcmToken string,
	eventTime time.Time, cmd command.PushUserTaxiCallCommand) (value.Notification, error) {
	messageTitle := "운행 취소"
	messageBody := "기사님이 택시 운행을 취소하였습니다."

	data := map[string]string{
		"taxiCallRequestId": cmd.TaxiCallRequestId,
		"taxiCallState":     cmd.TaxiCallState,
	}

	return value.NewNotification(fcmToken, value.NotificationCategory_Taxicall, messageTitle, messageBody, data), nil
}
