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
		"updateTime":             cmd.UpdateTime.Format(time.RFC3339),
	}

	return value.NewNotification(fcmToken, value.NotificationCategory_Taxicall, "", "", "", data), nil
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

	messageTitle := "배차 완료, 기사님이 이동중입니다."
	messageBody := fmt.Sprintf("예상운임 %d원 (서비스 이용료 %d원). "+
		"1분 경과후 취소 또는 기사님 도착 후 3분이내 연락두절, 미탑승시 취소수수료가 부과될 수 있습니다.",
		cmd.RequestBasePrice, cmd.AdditionalPrice)

	data := map[string]string{
		"taxiCallRequestId":   cmd.TaxiCallRequestId,
		"taxiCallState":       cmd.TaxiCallState,
		"driverId":            cmd.DriverId,
		"driverName":          driver.FullName(),
		"driverPhone":         driver.Phone,
		"driverTaxiCategory":  driver.CarProfile.TaxiCategory,
		"driverCarNumber":     driver.CarProfile.CarNumber,
		"driverCarModel":      driver.CarProfile.CarModel,
		"requestBasePrice":    fmt.Sprint(cmd.RequestBasePrice),
		"additionalPrice":     fmt.Sprint(cmd.AdditionalPrice),
		"usedPoint":           fmt.Sprint(cmd.UsedPoint),
		"driverLatitude":      fmt.Sprint(cmd.DriverLocation.Latitude),
		"driverLongitude":     fmt.Sprint(cmd.DriverLocation.Longitude),
		"departureLatitude":   fmt.Sprint(cmd.Departure.Point.Latitude),
		"departureLongitude":  fmt.Sprint(cmd.Departure.Point.Longitude),
		"arrivalLatitude":     fmt.Sprint(cmd.Arrival.Point.Latitude),
		"arrivalLongitude":    fmt.Sprint(cmd.Arrival.Point.Longitude),
		"toDepartureDistance": fmt.Sprint(cmd.ToDepartureDistance),
		"toDepartureETA":      fmt.Sprint(cmd.ToDepartureETA.Nanoseconds()),
		"toArrivalDistance":   fmt.Sprint(cmd.ToArrivalDistance),
		"toArrivalETA":        fmt.Sprint(cmd.ToArrivalETA.Nanoseconds()),
		"updateTime":          cmd.UpdateTime.Format(time.RFC3339),
	}

	return value.NewNotification(fcmToken, value.NotificationCategory_Taxicall, messageTitle, messageBody, "", data), nil
}

func (t taxiCallPushApp) handleUserTaxiCallRequestFailed(ctx context.Context, fcmToken string,
	eventTime time.Time, cmd command.PushUserTaxiCallCommand) (value.Notification, error) {

	messageTitle := "배차 실패, 다시 시도해주세요."
	messageBody := "호출을 수락하는 택시를 찾지 못했습니다. 이용료 상한을 조정하여 다시 시도해주세요."

	data := map[string]string{
		"taxiCallRequestId": cmd.TaxiCallRequestId,
		"taxiCallState":     cmd.TaxiCallState,
	}

	return value.NewNotification(fcmToken, value.NotificationCategory_Taxicall, messageTitle, messageBody, "", data), nil
}

func (t taxiCallPushApp) handleUserTaxiCallDriverToArrival(ctx context.Context, fcmToken string,
	eventTime time.Time, cmd command.PushUserTaxiCallCommand) (value.Notification, error) {
	messageTitle := "운행 시작, 목적지로 이동합니다."
	messageBody := "나에게 딱 맞는 택시와 목적지로 이동합니다. " +
		"지금 함께하고 계신 기사님은 누군가의 소중한 가족입니다. " +
		"승객과 기사님 서로가 배려하는 따뜻한 이동을 만드는데 동참해주세요."

	data := map[string]string{
		"taxiCallRequestId": cmd.TaxiCallRequestId,
		"taxiCallState":     cmd.TaxiCallState,
		"toArrivalDistance": fmt.Sprint(cmd.ToArrivalDistance),
		"toArrivalETA":      fmt.Sprint(cmd.ToArrivalETA.Nanoseconds()),
		"updateTime":        cmd.UpdateTime.Format(time.RFC3339),
	}

	return value.NewNotification(fcmToken, value.NotificationCategory_Taxicall, messageTitle, messageBody, "", data), nil
}

func (t taxiCallPushApp) handleUserTaxiCallRequestDone(ctx context.Context, fcmToken string,
	eventTime time.Time, cmd command.PushUserTaxiCallCommand) (value.Notification, error) {

	messageTitle := "운행 완료, 목적지에 도착하였습니다."
	messageBody := "나에게 딱 맞는 택시와 목적지에 도착하였습니다. 오늘도 타코택시와 함께해주셔서 감사합니다."

	data := map[string]string{
		"taxiCallRequestId": cmd.TaxiCallRequestId,
		"taxiCallState":     cmd.TaxiCallState,
		"basePrice":         fmt.Sprint(cmd.BasePrice),
		"additionalPrice":   fmt.Sprint(cmd.AdditionalPrice),
		"updateTime":        cmd.UpdateTime.Format(time.RFC3339),
	}

	return value.NewNotification(fcmToken, value.NotificationCategory_Taxicall, messageTitle, messageBody, "", data), nil
}

func (t taxiCallPushApp) handleDriverTaxiCallRequestCanceled(ctx context.Context, fcmToken string,
	eventTime time.Time, cmd command.PushUserTaxiCallCommand) (value.Notification, error) {
	messageTitle := "운행 취소, 기사님이 운행을 취소하였습니다."
	messageBody := "기사님이 택시 운행을 취소하였습니다. " +
		"기사님 도착후 연락두절, 장시간 미탑승 등 귀책사유가 있는 경우 취소 수수료가 발생할 수 있습니다."

	data := map[string]string{
		"taxiCallRequestId": cmd.TaxiCallRequestId,
		"taxiCallState":     cmd.TaxiCallState,
		"updateTime":        cmd.UpdateTime.Format(time.RFC3339),
	}

	return value.NewNotification(fcmToken, value.NotificationCategory_Taxicall, messageTitle, messageBody, "", data), nil
}

func (t taxiCallPushApp) handleDriverNotAvailable(ctx context.Context, fcmToken string,
	eventTime time.Time, cmd command.PushUserTaxiCallCommand) (value.Notification, error) {

	messageTitle := "배차 실패, 주변에 호출 가능한 택시가 없습니다"
	messageBody := "현재 주변에 호출 가능한 택시가 없습니다. 다시 시도해 주세요."

	data := map[string]string{
		"taxiCallRequestId": cmd.TaxiCallRequestId,
		"taxiCallState":     cmd.TaxiCallState,
	}

	return value.NewNotification(fcmToken, value.NotificationCategory_Taxicall, messageTitle, messageBody, "", data), nil
}

func (t taxiCallPushApp) handleDriverMockRequestAccepted(ctx context.Context, fcmToken string,
	eventTime time.Time, cmd command.PushUserTaxiCallCommand) (value.Notification, error) {

	messageTitle := "Mock Account 배차 요청 수락됨"
	messageBody := "수락됨"

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

	data := map[string]string{
		"taxiCallRequestId":   cmd.TaxiCallRequestId,
		"taxiCallState":       cmd.TaxiCallState,
		"driverName":          driver.FullName(),
		"driverId":            cmd.DriverId,
		"driverPhone":         driver.Phone,
		"driverTaxiCategory":  driver.CarProfile.TaxiCategory,
		"driverCarNumber":     driver.CarProfile.CarNumber,
		"driverCarModel":      driver.CarProfile.CarModel,
		"requestBasePrice":    fmt.Sprint(cmd.RequestBasePrice),
		"additionalPrice":     fmt.Sprint(cmd.AdditionalPrice),
		"usedPoint":           fmt.Sprint(cmd.UsedPoint),
		"driverLatitude":      fmt.Sprint(cmd.DriverLocation.Latitude),
		"driverLongitude":     fmt.Sprint(cmd.DriverLocation.Longitude),
		"departureLatitude":   fmt.Sprint(cmd.Departure.Point.Latitude),
		"departureLongitude":  fmt.Sprint(cmd.Departure.Point.Longitude),
		"arrivalLatitude":     fmt.Sprint(cmd.Arrival.Point.Latitude),
		"arrivalLongitude":    fmt.Sprint(cmd.Arrival.Point.Longitude),
		"toDepartureDistance": fmt.Sprint(cmd.ToDepartureDistance),
		"toDepartureETA":      fmt.Sprint(cmd.ToDepartureETA.Nanoseconds()),
		"toArrivalDistance":   fmt.Sprint(cmd.ToArrivalDistance),
		"toArrivalETA":        fmt.Sprint(cmd.ToArrivalETA.Nanoseconds()),
		"updateTime":          cmd.UpdateTime.Format(time.RFC3339),
	}

	return value.NewNotification(fcmToken, value.NotificationCategory_Taxicall, messageTitle, messageBody, "", data), nil
}
