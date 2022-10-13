package push

import (
	"context"
	"fmt"

	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/repository"
	"github.com/taco-labs/taco/go/service"
	"github.com/uptrace/bun"
	"golang.org/x/sync/errgroup"
)

type taxiCallPushApp struct {
	app.Transactor
	repository struct {
		user   repository.UserRepository
		driver repository.DriverRepository
	}
	service struct {
		route        service.MapRouteService
		notification service.NotificationService
	}
}

func (t taxiCallPushApp) DistributeTaxiCallRequest(ctx context.Context,
	taxiCallRequest entity.TaxiCallRequest,
	taxiCallTicket entity.TaxiCallTicket,
	driverContexts []entity.DriverTaxiCallContext) error {

	type driverContextWithFcmToken struct {
		entity.DriverTaxiCallContext
		AppFcmToken string
	}

	driverContextWithFcmTokens := make([]driverContextWithFcmToken, 0, len(driverContexts))

	err := t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		for _, driverContext := range driverContexts {
			dr, err := t.repository.driver.FindById(ctx, i, driverContext.DriverId)
			if err != nil {
				return fmt.Errorf("service.TaxiCallPush.DistributeTaxiCallRequest: error while get driver: %w", err)
			}
			driverContextWithFcmTokens = append(driverContextWithFcmTokens, driverContextWithFcmToken{
				DriverTaxiCallContext: driverContext,
				AppFcmToken:           dr.AppFcmToken,
			})
		}
		return nil
	})

	if err != nil {
		return err
	}

	group, nctx := errgroup.WithContext(ctx)
	resultChan := make(chan value.TitledNotification, len(driverContextWithFcmTokens))
	notifications := make([]value.Notification, 0, len(driverContextWithFcmTokens))

	for _, driverContext := range driverContextWithFcmTokens {
		dctx := driverContext
		group.Go(func() error {
			routeBetweenDeparture, err := t.service.route.GetRoute(nctx, dctx.Location, taxiCallRequest.Departure.Point)
			if err != nil {
				return fmt.Errorf("service.TaxiCallPush.DistributeTaxiCallRequest: error while get route between driver location and departure: %w", err)
			}

			// TODO (taekyeom) fix it
			message := value.NotificationMessage{
				Title: fmt.Sprintf("배차 요청 (약 %d분)", int(routeBetweenDeparture.ETA.Minutes())),
				Body:  fmt.Sprintf("%s (추가 요금 %d)", taxiCallRequest.Departure.Address.AddressName, taxiCallTicket.AdditionalPrice),
			}

			data := map[string]string{
				"taxiCallRequestId":           taxiCallRequest.Id,
				"taxiCallState":               string(taxiCallRequest.CurrentState),
				"taxiCallTicketId":            taxiCallTicket.Id,
				"requestBasePrice":            fmt.Sprint(taxiCallRequest.RequestBasePrice),
				"additionalPrice":             fmt.Sprint(taxiCallTicket.AdditionalPrice),
				"toDepartureDistance":         fmt.Sprint(routeBetweenDeparture.Distance),
				"toDepartureETA":              fmt.Sprint(routeBetweenDeparture.ETA),
				"departureAdressRegionDepth1": taxiCallRequest.Departure.Address.RegionDepth1,
				"departureAdressRegionDepth2": taxiCallRequest.Departure.Address.RegionDepth2,
				"departureAdressRegionDepth3": taxiCallRequest.Departure.Address.RegionDepth3,
				"departureMainAddressNo":      taxiCallRequest.Departure.Address.MainAddressNo,
				"departureSubAddressNo":       taxiCallRequest.Departure.Address.SubAddressNo,
				"arrivalAdressRegionDepth1":   taxiCallRequest.Arrival.Address.RegionDepth1,
				"arrivalAdressRegionDepth2":   taxiCallRequest.Arrival.Address.RegionDepth2,
				"arrivalAdressRegionDepth3":   taxiCallRequest.Arrival.Address.RegionDepth3,
				"arrivalMainAddressNo":        taxiCallRequest.Arrival.Address.MainAddressNo,
				"arrivalSubAddressNo":         taxiCallRequest.Arrival.Address.SubAddressNo,
			}

			notification := value.NewTitledNotification(dctx.AppFcmToken, message, data)
			resultChan <- notification
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return err
	}

	for i := 0; i < len(resultChan); i++ {
		notifications = append(notifications, <-resultChan)
	}

	bulkSendPushMessage(t.service.notification, notifications)

	return nil
}

func (t taxiCallPushApp) SendTaxiCallRequestAccept(ctx context.Context,
	taxiCallRequest entity.TaxiCallRequest,
	driverContext entity.DriverTaxiCallContext) error {

	var user entity.User
	err := t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		u, err := t.repository.user.FindById(ctx, i, taxiCallRequest.UserId)
		if err != nil {
			return fmt.Errorf("app.push.SendDriverDriverTaxiCallRequestAccept: error while find user: %w", err)
		}
		user = u
		return nil
	})

	if err != nil {
		return err
	}

	routeBetweenDeparture, err := t.service.route.GetRoute(ctx, driverContext.Location, taxiCallRequest.Departure.Point)
	if err != nil {
		return fmt.Errorf("app.push.SendDriverDriverTaxiCallRequestAccept: error while get route between driver location and departure: %w", err)
	}

	message := value.NotificationMessage{
		Title: fmt.Sprintf("배차 완료 (약 %d분)", int(routeBetweenDeparture.ETA.Minutes())),
		Body:  fmt.Sprintf("%s (추가 요금 %d)", taxiCallRequest.Departure.Address.AddressName, taxiCallRequest.AdditionalPrice),
	}

	data := map[string]string{
		"taxiCallRequestId":     taxiCallRequest.Id,
		"taxiCallState":         string(taxiCallRequest.CurrentState),
		"driverId":              taxiCallRequest.DriverId.String, // TODO (driver informations)
		"requestBasePrice":      fmt.Sprint(taxiCallRequest.RequestBasePrice),
		"additionalPrice":       fmt.Sprint(taxiCallRequest.AdditionalPrice),
		"toDepartureDistance":   fmt.Sprint(routeBetweenDeparture.Distance),
		"whenDriverToDeparture": fmt.Sprint(taxiCallRequest.UpdateTime.Add(routeBetweenDeparture.ETA)),
	}

	notification := value.NewTitledNotification(user.AppFcmToken, message, data)
	sendPushMessage(t.service.notification, notification)

	return nil
}

func (t taxiCallPushApp) SendTaxiCallRequestDone(ctx context.Context,
	taxiCallRequest entity.TaxiCallRequest) error {

	var user entity.User
	err := t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		u, err := t.repository.user.FindById(ctx, i, taxiCallRequest.UserId)
		if err != nil {
			return fmt.Errorf("app.push.SendTaxiCallRequestDone: error while find user: %w", err)
		}
		user = u
		return nil
	})

	if err != nil {
		return err
	}

	message := value.NotificationMessage{
		Title: "운행 완료",
		Body:  "택시 운행이 완료되었습니다.",
	}

	data := map[string]string{
		"taxiCallRequestId": taxiCallRequest.Id,
		"taxiCallState":     string(taxiCallRequest.CurrentState),
		"basePrice":         fmt.Sprint(taxiCallRequest.BasePrice),
		"additionalPrice":   fmt.Sprint(taxiCallRequest.AdditionalPrice),
	}

	notification := value.NewTitledNotification(user.AppFcmToken, message, data)
	sendPushMessage(t.service.notification, notification)

	return nil
}

func (t taxiCallPushApp) SendTaxiCallRequestFailure(ctx context.Context,
	taxiCallRequest entity.TaxiCallRequest) error {

	var user entity.User
	err := t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		u, err := t.repository.user.FindById(ctx, i, taxiCallRequest.UserId)
		if err != nil {
			return fmt.Errorf("app.push.SendTaxiCallRequestFailure: error while find user: %w", err)
		}
		user = u
		return nil
	})

	if err != nil {
		return err
	}

	message := value.NotificationMessage{
		Title: "배차 실패",
		Body:  "택시 배차에 실패했습니다",
	}

	data := map[string]string{
		"taxiCallRequestId": taxiCallRequest.Id,
		"taxiCallState":     string(taxiCallRequest.CurrentState),
	}

	notification := value.NewTitledNotification(user.AppFcmToken, message, data)
	sendPushMessage(t.service.notification, notification)

	return nil
}

func (t taxiCallPushApp) SendTaxiCallRequestProgress(ctx context.Context,
	taxiCallRequest entity.TaxiCallRequest, ticket entity.TaxiCallTicket) error {
	var user entity.User
	err := t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		u, err := t.repository.user.FindById(ctx, i, taxiCallRequest.UserId)
		if err != nil {
			return fmt.Errorf("app.push.SendTaxiCallRequestProgress: error while find user: %w", err)
		}
		user = u
		return nil
	})

	if err != nil {
		return err
	}

	data := map[string]string{
		"taxiCallRequestId":      taxiCallRequest.Id,
		"taxiCallState":          string(taxiCallRequest.CurrentState),
		"currentAdditionalPrice": fmt.Sprint(ticket.AdditionalPrice),
		"searchRangeInMinutes":   fmt.Sprint(ticket.GetRadiusMinutes()),
	}

	notification := value.NewDataOnlyNotification(user.AppFcmToken, data)
	sendPushMessage(t.service.notification, notification)

	return nil
}

func sendPushMessage(svc service.NotificationService, notification value.Notification) {
	go func() {
		ctx := context.Background()
		err := svc.SendNotification(ctx, notification)
		if err != nil {
			// TODO(taekyeom) Logging
			fmt.Println("Erorr while send notification: ", err)
		}
	}()
}

func bulkSendPushMessage(svc service.NotificationService, notifications []value.Notification) {
	fmt.Println("Send Messages: ", notifications)
	go func() {
		ctx := context.Background()
		err := svc.BulkSendNotification(ctx, notifications)
		if err != nil {
			// TODO(taekyeom) Logging
			fmt.Println("Erorr while bulk send notification: ", err)
		}
	}()
}
