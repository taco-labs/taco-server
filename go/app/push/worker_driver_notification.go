package push

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/event/command"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/uptrace/bun"
)

func (t taxiCallPushApp) handleDriverTaxiCallRequestTicketDistribution(ctx context.Context, fcmToken string,
	eventTime time.Time, cmd command.PushDriverTaxiCallCommand) (value.Notification, error) {

	var user entity.User
	err := t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
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

	data := map[string]string{
		"taxiCallRequestId":            cmd.TaxiCallRequestId,
		"taxiCallState":                cmd.TaxiCallState,
		"taxiCallTicketId":             cmd.TaxiCallTicketId,
		"ticketAttempt":                fmt.Sprint(cmd.Attempt),
		"requestBasePrice":             fmt.Sprint(cmd.RequestBasePrice),
		"additionalPrice":              fmt.Sprint(cmd.AdditionalPrice),
		"toArrivalDistance":            fmt.Sprint(cmd.ToArrivalDistance),
		"toArrivalETA":                 fmt.Sprint(cmd.ToArrivalETA.Nanoseconds()),
		"toDepartureDistance":          fmt.Sprint(cmd.ToDepartureDistance),
		"departureLatitude":            fmt.Sprint(cmd.Departure.Point.Latitude),
		"departureLongitude":           fmt.Sprint(cmd.Departure.Point.Longitude),
		"departureAddressRegionDepth1": cmd.Departure.Address.RegionDepth1,
		"departureAddressRegionDepth2": cmd.Departure.Address.RegionDepth2,
		"departureAddressRegionDepth3": cmd.Departure.Address.RegionDepth3,
		"departureMainAddressNo":       cmd.Departure.Address.MainAddressNo,
		"departureSubAddressNo":        cmd.Departure.Address.SubAddressNo,
		"departureBuildingName":        cmd.Departure.Address.BuildingName,
		"arrivalLatitude":              fmt.Sprint(cmd.Arrival.Point.Latitude),
		"arrivalLongitude":             fmt.Sprint(cmd.Arrival.Point.Longitude),
		"arrivalAddressRegionDepth1":   cmd.Arrival.Address.RegionDepth1,
		"arrivalAddressRegionDepth2":   cmd.Arrival.Address.RegionDepth2,
		"arrivalAddressRegionDepth3":   cmd.Arrival.Address.RegionDepth3,
		"arrivalMainAddressNo":         cmd.Arrival.Address.MainAddressNo,
		"arrivalSubAddressNo":          cmd.Arrival.Address.SubAddressNo,
		"arrivalBuildingName":          cmd.Arrival.Address.BuildingName,
		"userId":                       cmd.UserId,
		"userPhone":                    user.Phone,
		"tags":                         strings.Join(cmd.Tags, ","),
		"userTag":                      cmd.UserTag,
		"updateTime":                   cmd.UpdateTime.Format(time.RFC3339),
	}

	return value.NewNotification(fcmToken, value.NotificationCategory_Taxicall, "", "", "taxicallRequestDistribution", data), nil
}

func (t taxiCallPushApp) handleUserTaxiCallRequestCanceled(ctx context.Context, fcmToken string,
	eventTime time.Time, cmd command.PushDriverTaxiCallCommand) (value.Notification, error) {

	messageTitle := "운행 취소, 승객이 호출을 취소하였습니다."
	messageBody := "승객이 호출을 취소하였습니다. 취소수수료 부과 대상인 경우 적립된 타코포인트는 정산관리 메뉴에서 확인 가능합니다."

	data := map[string]string{
		"taxiCallRequestId": cmd.TaxiCallRequestId,
		"taxiCallState":     cmd.TaxiCallState,
		"updateTime":        cmd.UpdateTime.Format(time.RFC3339),
	}

	return value.NewNotification(fcmToken, value.NotificationCategory_Taxicall, messageTitle, messageBody, "", data), nil
}
