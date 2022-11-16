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

func (t taxiCallPushApp) handleDriverTaxiCallRequestTicketDistribution(ctx context.Context, fcmToken string,
	eventTime time.Time, cmd command.PushDriverTaxiCallCommand) (value.Notification, error) {

	routeBetweenDeparture, err := t.service.route.GetRoute(ctx, cmd.DriverLocation, cmd.Departure.Point)
	if err != nil {
		return value.Notification{},
			fmt.Errorf("service.TaxiCallPush.handleDriverTaxiCallRequestTicketDistribution: error while get route between driver location and departure: %w", err)
	}

	var user entity.User
	err = t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		u, err := t.service.userGetter.GetUser(ctx, cmd.UserId)
		if err != nil {
			return fmt.Errorf("service.TaxiCallPush.handleDriverTaxiCallRequestTicketDistribution: error while get user: %w", err)
		}
		user = u
		return nil
	})
	if err != nil {
		return value.Notification{}, err
	}

	message := value.NotificationMessage{
		Title: fmt.Sprintf("배차 요청 (약 %d분)", int(routeBetweenDeparture.ETA.Minutes())),
		Body:  fmt.Sprintf("%s (추가 요금 %d)", cmd.Departure.Address.AddressName, cmd.AdditionalPrice),
	}

	data := map[string]string{
		"taxiCallRequestId":           cmd.TaxiCallRequestId,
		"taxiCallState":               cmd.TaxiCallState,
		"taxiCallTicketId":            cmd.TaxiCallTicketId,
		"requestBasePrice":            fmt.Sprint(cmd.RequestBasePrice),
		"additionalPrice":             fmt.Sprint(cmd.AdditionalPrice),
		"toDepartureDistance":         fmt.Sprint(routeBetweenDeparture.Distance),
		"toDepartureETA":              fmt.Sprint(routeBetweenDeparture.ETA),
		"departureLatitude":           fmt.Sprint(cmd.Departure.Point.Latitude),
		"departureLongitude":          fmt.Sprint(cmd.Departure.Point.Longitude),
		"departureAdressRegionDepth1": cmd.Departure.Address.RegionDepth1,
		"departureAdressRegionDepth2": cmd.Departure.Address.RegionDepth2,
		"departureAdressRegionDepth3": cmd.Departure.Address.RegionDepth3,
		"departureMainAddressNo":      cmd.Departure.Address.MainAddressNo,
		"departureSubAddressNo":       cmd.Departure.Address.SubAddressNo,
		"arrivalLatitude":             fmt.Sprint(cmd.Arrival.Point.Latitude),
		"arrivalLongitude":            fmt.Sprint(cmd.Arrival.Point.Longitude),
		"arrivalAdressRegionDepth1":   cmd.Arrival.Address.RegionDepth1,
		"arrivalAdressRegionDepth2":   cmd.Arrival.Address.RegionDepth2,
		"arrivalAdressRegionDepth3":   cmd.Arrival.Address.RegionDepth3,
		"arrivalMainAddressNo":        cmd.Arrival.Address.MainAddressNo,
		"arrivalSubAddressNo":         cmd.Arrival.Address.SubAddressNo,
		"userId":                      cmd.UserId,
		"userPhone":                   user.Phone,
	}

	return value.Notification{
		Principal: fcmToken,
		Message:   message,
		Data:      data,
	}, nil
}

func (t taxiCallPushApp) handleUserTaxiCallRequestCanceled(ctx context.Context, fcmToken string,
	eventTime time.Time, cmd command.PushDriverTaxiCallCommand) (value.Notification, error) {
	message := value.NotificationMessage{
		Title: "운행 취소",
		Body:  "승객이 택시 운행을 취소하였습니다.",
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
