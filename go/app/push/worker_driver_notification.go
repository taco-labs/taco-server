package push

import (
	"context"
	"fmt"
	"time"

	"github.com/taco-labs/taco/go/domain/event/command"
	"github.com/taco-labs/taco/go/domain/value"
)

func (t taxiCallPushApp) handleDriverTaxiCallRequestTicketDistribution(ctx context.Context, fcmToken string,
	eventTime time.Time, cmd command.DriverTaxiCallNotificationCommand) (value.Notification, error) {

	routeBetweenDeparture, err := t.service.route.GetRoute(ctx, cmd.DriverLocation, cmd.Departure.Point)
	if err != nil {
		return value.Notification{},
			fmt.Errorf("service.TaxiCallPush.handleDriverTaxiCallRequestTicketDistribution: error while get route between driver location and departure: %w", err)
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
		"departureAdressRegionDepth1": cmd.Departure.Address.RegionDepth1,
		"departureAdressRegionDepth2": cmd.Departure.Address.RegionDepth2,
		"departureAdressRegionDepth3": cmd.Departure.Address.RegionDepth3,
		"departureMainAddressNo":      cmd.Departure.Address.MainAddressNo,
		"departureSubAddressNo":       cmd.Departure.Address.SubAddressNo,
		"arrivalAdressRegionDepth1":   cmd.Arrival.Address.RegionDepth1,
		"arrivalAdressRegionDepth2":   cmd.Arrival.Address.RegionDepth2,
		"arrivalAdressRegionDepth3":   cmd.Arrival.Address.RegionDepth3,
		"arrivalMainAddressNo":        cmd.Arrival.Address.MainAddressNo,
		"arrivalSubAddressNo":         cmd.Arrival.Address.SubAddressNo,
	}

	return value.Notification{
		Principal: fcmToken,
		Message:   message,
		Data:      data,
	}, nil
}
