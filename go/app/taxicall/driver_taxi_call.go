package taxicall

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/taco-labs/taco/go/common/analytics"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/event/command"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/domain/value/enum"
	"github.com/taco-labs/taco/go/utils"
	"github.com/taco-labs/taco/go/utils/slices"
	"github.com/uptrace/bun"
)

func (t taxicallApp) ActivateDriverContext(ctx context.Context, driverId string) error {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	return t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		taxiCallContext, err := t.repository.taxiCallRequest.GetDriverTaxiCallContext(ctx, i, driverId)
		if err != nil && !errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("app.taxiCall.ActivateDriverContext: error while get last call request: %w", err)
		}
		if errors.Is(err, value.ErrNotFound) {
			taxiCallContext = entity.NewEmptyDriverTaxiCallContext(driverId, true, requestTime)
		}
		taxiCallContext.CanReceive = true

		if err := t.repository.taxiCallRequest.UpsertDriverTaxiCallContext(ctx, i, taxiCallContext); err != nil {
			return fmt.Errorf("app.taxiCall.ActivateDriverContext: error while upsert driver taxi call context: %w", err)
		}

		return nil
	})
}

func (t taxicallApp) DeactivateDriverContext(ctx context.Context, driverId string) error {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	return t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		lastTaxiCallRequest, err := t.repository.taxiCallRequest.GetLatestByDriverId(ctx, i, driverId)
		if err != nil && !errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("app.taxiCall.DeactivateDriverContext: error while get last call request: %w", err)
		}
		if lastTaxiCallRequest.CurrentState.Active() {
			return fmt.Errorf("app.taxiCall.DeactivateDriverContext: active taxi call request exists: %w", value.ErrActiveTaxiCallRequestExists)
		}

		taxiCallContext, err := t.repository.taxiCallRequest.GetDriverTaxiCallContext(ctx, i, driverId)
		if err != nil && !errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("app.taxiCall.DeactivateDriverContext: error while get last call request: %w", err)
		}
		if errors.Is(err, value.ErrNotFound) {
			taxiCallContext = entity.NewEmptyDriverTaxiCallContext(driverId, false, requestTime)
		}
		taxiCallContext.CanReceive = false

		if err := t.repository.taxiCallRequest.UpsertDriverTaxiCallContext(ctx, i, taxiCallContext); err != nil {
			return fmt.Errorf("app.taxiCall.DeactivateDriverContext: error while upsert driver taxi call context: %w", err)
		}

		return nil
	})
}

func (t taxicallApp) UpdateDriverLocation(ctx context.Context, req request.DriverLocationUpdateRequest) error {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	return t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		driverLocationDto := entity.DriverLocation{
			DriverId: req.DriverId,
			Location: value.Point{
				Latitude:  req.Latitude,
				Longitude: req.Longitude,
			},
			UpdateTime: requestTime,
		}

		if err := t.repository.driverLocation.Upsert(ctx, i, driverLocationDto); err != nil {
			return fmt.Errorf("app.taxiCall.UpdateDriverLocation: error while update driver location:\n%w", err)
		}

		return nil
	})
}

func (t taxicallApp) ListDriverTaxiCallRequest(ctx context.Context, req request.ListDriverTaxiCallRequest) ([]entity.TaxiCallRequest, string, error) {
	var taxiCallRequests []entity.TaxiCallRequest
	var err error
	var pageToken string

	err = t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		taxiCallRequests, pageToken, err = t.repository.taxiCallRequest.ListByDriverId(ctx, i, req.DriverId, req.PageToken, req.Count)
		if err != nil {
			return fmt.Errorf("app.taxiCall.ListDriverTaxiCallRequest: error while get taxi call requests:\n%w", err)
		}

		err := slices.ForeachErrRef(taxiCallRequests, func(i *entity.TaxiCallRequest) error {
			tags, err := slices.MapErr(i.TagIds, value.GetTagById)
			if err != nil {
				return err
			}
			i.Tags = tags
			return nil
		})
		if err != nil {
			return fmt.Errorf("app.taxiCall.ListDriverTaxiCallRequest: error while get tags: %w", err)
		}

		return nil
	})

	if err != nil {
		return []entity.TaxiCallRequest{}, "", nil
	}

	return taxiCallRequests, pageToken, nil
}

func (t taxicallApp) LatestDriverTaxiCallRequest(ctx context.Context, driverId string) (entity.DriverLatestTaxiCallRequest, error) {
	var latestTaxiCallRequest entity.DriverLatestTaxiCallRequest
	var err error

	err = t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		taxiCallRequest, err := t.repository.taxiCallRequest.GetLatestByDriverId(ctx, i, driverId)
		if err != nil {
			return fmt.Errorf("app.driver.LatestDriverTaxiCallRequest: error while get latest taxi call:\n%w", err)
		}

		tags, err := slices.MapErr(taxiCallRequest.TagIds, value.GetTagById)
		if err != nil {
			return fmt.Errorf("app.taxiCall.LatestDriverTaxiCallRequest: error while get tags: %w", err)
		}

		taxiCallRequest.Tags = tags

		user, err := t.service.userGetter.GetUser(ctx, taxiCallRequest.UserId)
		if err != nil {
			return fmt.Errorf("app.taxiCall.LatestDriverTaxiCallRequest: error while get user: %w", err)
		}

		latestTaxiCallRequest = entity.DriverLatestTaxiCallRequest{
			TaxiCallRequest: taxiCallRequest,
			UserPhone:       user.Phone,
		}

		return nil
	})

	if err != nil {
		return entity.DriverLatestTaxiCallRequest{}, err
	}

	return latestTaxiCallRequest, nil
}

// TODO (taekyeom) Remove it later!!
func (t taxicallApp) ForceAcceptTaxiCallRequest(ctx context.Context, driverId, callRequestId string) error {
	return t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		ticket, err := t.repository.taxiCallRequest.GetLatestTicketByRequestId(ctx, i, callRequestId)
		if err != nil {
			return fmt.Errorf("app.taxiCall.ForceAcceptTaxiCallRequest: error while find latest ticket by request id: %w", err)
		}

		// TODO(taeykeom) Do we need check on duty & last call request?
		driverTaxiCallContext, err := t.repository.taxiCallRequest.GetDriverTaxiCallContext(ctx, i, driverId)
		if err != nil {
			return fmt.Errorf("app.taxiCall.ForceAcceptTaxiCallRequest: error while get taxi call context:%w", err)
		}

		driverTaxiCallContext.CanReceive = false
		driverTaxiCallContext.LastReceivedRequestTicket = ticket.TicketId
		if err := t.repository.taxiCallRequest.UpsertDriverTaxiCallContext(ctx, i, driverTaxiCallContext); err != nil {
			return fmt.Errorf("app.taxiCall.ForceAcceptTaxiCallRequest: error while upsert taxi call context: %w", value.ErrInvalidOperation)
		}

		return t.AcceptTaxiCallRequest(ctx, driverId, ticket.TicketId)
	})
}

// TODO (taekyeom) Add route between driver location & departure?
func (t taxicallApp) AcceptTaxiCallRequest(ctx context.Context, driverId string, ticketId string) error {
	requestTime := utils.GetRequestTimeOrNow(ctx)
	var taxiCallRequest entity.TaxiCallRequest
	var driverTaxiCallContext entity.DriverTaxiCallContext
	var err error

	err = t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		// TODO(taeykeom) Do we need check on duty & last call request?
		driverTaxiCallContext, err = t.repository.taxiCallRequest.GetDriverTaxiCallContext(ctx, i, driverId)
		if err != nil {
			return fmt.Errorf("app.taxiCall.AcceptTaxiCallRequest: error while get taxi call context:%w", err)
		}
		if driverTaxiCallContext.LastReceivedRequestTicket != ticketId {
			return fmt.Errorf("app.taxiCall.AcceptTaxiCallRequest: invalid ticket id: %w", value.ErrInvalidOperation)
		}

		driverTaxiCallContext.CanReceive = false
		if err := t.repository.taxiCallRequest.UpsertDriverTaxiCallContext(ctx, i, driverTaxiCallContext); err != nil {
			return fmt.Errorf("app.taxiCall.AcceptTaxiCallRequest: error while upsert taxi call context: %w", value.ErrInvalidOperation)
		}

		receivedTicket, err := t.repository.taxiCallRequest.GetTicketById(ctx, i, driverTaxiCallContext.LastReceivedRequestTicket)
		if err != nil {
			return fmt.Errorf("app.taxiCall.AcceptTaxiCallRequest: error while get taxi call ticket:%w", err)
		}

		taxiCallRequest, err = t.repository.taxiCallRequest.GetById(ctx, i, receivedTicket.TaxiCallRequestId)
		if err != nil {
			return fmt.Errorf("app.taxiCall.AcceptTaxiCallRequest: error while get taxi call request:%w", err)
		}
		if !taxiCallRequest.CurrentState.Requested() {
			return fmt.Errorf("app.taxiCall.AcceptTaxiCallRequest: already expired taxi call request:%w", value.ErrAlreadyExpiredCallRequest)
		}

		// TODO(taekyeom) actualTicket과 현재 actualTicket이 다른 경우, 돈을 더 받는것도 괜찮을까?
		actualTicket, err := t.repository.taxiCallRequest.GetLatestTicketByRequestId(ctx, i, taxiCallRequest.Id)
		if err != nil {
			return fmt.Errorf("app.taxiCall.AcceptTaxiCallRequest: error while get latest taxi call ticket:%w", err)
		}

		if err := taxiCallRequest.UpdateState(requestTime, enum.TaxiCallState_DRIVER_TO_DEPARTURE); err != nil {
			return fmt.Errorf("app.taxiCall.AcceptTaxiCallRequest: invalid state change:%w", err)
		}

		// TODO (taekyeom) seperate external api call
		routeToDeparture, err := t.service.route.GetRoute(ctx, driverTaxiCallContext.Location, taxiCallRequest.Departure.Point)
		if err != nil {
			return fmt.Errorf("app.taxiCall.AcceptTaxiCallRequest: error while get route between departure:%w", err)
		}

		taxiCallRequest.AdditionalPrice = actualTicket.AdditionalPrice
		taxiCallRequest.ToDepartureRoute = routeToDeparture
		taxiCallRequest.DriverId = sql.NullString{
			Valid:  true,
			String: driverId,
		}
		if err := t.repository.taxiCallRequest.Update(ctx, i, taxiCallRequest); err != nil {
			return fmt.Errorf("app.taxiCall.AcceptTaxiCallRequest: error while update taxi call request :%w", err)
		}

		processMessage := command.NewTaxiCallProgressCommand(
			taxiCallRequest.Id,
			taxiCallRequest.CurrentState,
			taxiCallRequest.UpdateTime,
			taxiCallRequest.UpdateTime,
		)

		if err := t.repository.event.BatchCreate(ctx, i, []entity.Event{processMessage}); err != nil {
			return fmt.Errorf("app.taxiCall.AcceptTaxiCallRequest: error while create taxi call process event: %w", err)
		}

		analytics.WriteAnalyticsLog(ctx, requestTime, analytics.LogType_DriverTaxiCallTicketAccept, analytics.DriverTaxiCallTicketAcceptPayload{
			DriverId:                        driverTaxiCallContext.DriverId,
			RequestUserId:                   taxiCallRequest.UserId,
			TaxiCallRequestId:               taxiCallRequest.Id,
			ReceivedTaxiCallRequestTicketId: receivedTicket.TicketId,
			ReceivedTicketAttempt:           receivedTicket.Attempt,
			ActualTaxiCallRequestTicketId:   actualTicket.TicketId,
			ActualTicketAttempt:             actualTicket.Attempt,
			RequestBasePrice:                taxiCallRequest.RequestBasePrice,
			AdditionalPrice:                 taxiCallRequest.AdditionalPrice,
			DriverLocation:                  driverTaxiCallContext.Location,
			ReceiveTime:                     driverTaxiCallContext.LastReceiveTime,
			TaxiCallRequestCreateTime:       taxiCallRequest.CreateTime,
		})

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (t taxicallApp) RejectTaxiCallRequest(ctx context.Context, driverId string, ticketId string) error {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	return t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		// TODO(taeykeom) Do we need check on duty & last call request?
		driverTaxiCallContext, err := t.repository.taxiCallRequest.GetDriverTaxiCallContext(ctx, i, driverId)
		if err != nil {
			return fmt.Errorf("app.taxiCall.RejectTaxiCallRequest: error while get taxi call context:%w", err)
		}
		if driverTaxiCallContext.LastReceivedRequestTicket != ticketId {
			return fmt.Errorf("app.taxiCall.RejectTaxiCallRequest: invalid ticket id: %w", value.ErrInvalidOperation)
		}

		driverTaxiCallContext.RejectedLastRequestTicket = true
		if err := t.repository.taxiCallRequest.UpsertDriverTaxiCallContext(ctx, i, driverTaxiCallContext); err != nil {
			return fmt.Errorf("app.taxiCall.RejectTaxiCallRequest: error while upsert taxi call context: %w", value.ErrInvalidOperation)
		}

		taxiCallTicket, err := t.repository.taxiCallRequest.GetTicketById(ctx, i, ticketId)
		if err != nil {
			return fmt.Errorf("app.taxiCall.RejectTaxiCallRequest: error while get ticket by id: %w", err)
		}

		taxiCallRequest, err := t.repository.taxiCallRequest.GetById(ctx, i, taxiCallTicket.TaxiCallRequestId)
		if err != nil {
			return fmt.Errorf("app.taxiCall.RejectTaxiCallRequest: error while get taxi call request by id: %w", err)
		}

		analytics.WriteAnalyticsLog(ctx, requestTime, analytics.LogType_DriverTaxiCallTicketReject, analytics.DriverTaxiCallTicketRejectPayload{
			DriverId:                  driverTaxiCallContext.DriverId,
			RequestUserId:             taxiCallRequest.UserId,
			TaxiCallRequestId:         taxiCallRequest.Id,
			TaxiCallRequestTicketId:   taxiCallTicket.TicketId,
			TicketAttempt:             taxiCallTicket.Attempt,
			RequestBasePrice:          taxiCallRequest.RequestBasePrice,
			AdditionalPrice:           taxiCallRequest.AdditionalPrice,
			DriverLocation:            driverTaxiCallContext.Location,
			ReceiveTime:               driverTaxiCallContext.LastReceiveTime,
			TaxiCallRequestCreateTime: taxiCallRequest.CreateTime,
		})

		return nil
	})
}

func (t taxicallApp) DriverCancelTaxiCallRequest(ctx context.Context, driverId string, taxiCallId string) error {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	return t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		taxiCallRequest, err := t.repository.taxiCallRequest.GetById(ctx, i, taxiCallId)
		if err != nil {
			return fmt.Errorf("app.taxCall.DriverCancelTaxiCall: error while get taxi call:%w", err)
		}

		if taxiCallRequest.DriverId.String != driverId {
			return fmt.Errorf("app.taxCall.DriverCancelTaxiCall: Invalid request:%w", value.ErrUnAuthorized)
		}

		taxiCallRequestAcceptTime := taxiCallRequest.UpdateTime

		if err = taxiCallRequest.UpdateState(requestTime, enum.TaxiCallState_DRIVER_CANCELLED); err != nil {
			return fmt.Errorf("app.taxCall.DriverCancelTaxiCall: error while cancel taxi call:%w", err)
		}

		if err = t.repository.taxiCallRequest.Update(ctx, i, taxiCallRequest); err != nil {
			return fmt.Errorf("app.taxCall.DriverCancelTaxiCall: error while update taxi call:%w", err)
		}

		driverTaxiCallContext, err := t.repository.taxiCallRequest.GetDriverTaxiCallContext(ctx, i, driverId)
		if err != nil {
			return fmt.Errorf("app.taxCall.DriverCancelTaxiCall: error while get taxi call context:%w", err)
		}
		driverTaxiCallContext.CanReceive = true
		driverTaxiCallContext.RejectedLastRequestTicket = true

		if err := t.repository.taxiCallRequest.UpsertDriverTaxiCallContext(ctx, i, driverTaxiCallContext); err != nil {
			return fmt.Errorf("app.taxiCall.DriverCancelTaxiCall: error while upsert taxi call context: %w", err)
		}

		processMessage := command.NewTaxiCallProgressCommand(
			taxiCallRequest.Id,
			taxiCallRequest.CurrentState,
			taxiCallRequest.UpdateTime,
			taxiCallRequest.UpdateTime,
		)

		if err := t.repository.event.BatchCreate(ctx, i, []entity.Event{processMessage}); err != nil {
			return fmt.Errorf("app.taxiCall.CreateTaxiCallRequest: error while create taxi call process event: %w", err)
		}

		analytics.WriteAnalyticsLog(ctx, requestTime, analytics.LogType_DriverTaxiCallCancel, analytics.DriverTaxiCancelPayload{
			DriverId:                  taxiCallRequest.DriverId.String,
			RequestUserId:             taxiCallRequest.UserId,
			TaxiCallRequestId:         taxiCallRequest.Id,
			DriverLocation:            driverTaxiCallContext.Location,
			TaxiCallRequestCreateTime: taxiCallRequest.CreateTime,
			AcceptTime:                taxiCallRequestAcceptTime,
		})

		return nil
	})
}

func (d taxicallApp) DriverToArrival(ctx context.Context, driverId string, callRequestId string) error {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	return d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		taxiCallRequest, err := d.repository.taxiCallRequest.GetById(ctx, i, callRequestId)
		if err != nil {
			return fmt.Errorf("app.taxiCall.DriverToArrival: error while get taxi call request: %w", err)
		}

		driverLocation, err := d.repository.driverLocation.GetByDriverId(ctx, i, driverId)
		if err != nil {
			return fmt.Errorf("app.taxiCall.DriverToArrival: error while get driver location: %w", err)
		}

		if taxiCallRequest.DriverId.String != driverId {
			return fmt.Errorf("app.taxiCall.DriverToArrival: unauthorized access: %w", value.ErrUnAuthorized)
		}

		taxiCallRequestAcceptTime := taxiCallRequest.UpdateTime

		if err := taxiCallRequest.UpdateState(requestTime, enum.TaxiCallState_DRIVER_TO_ARRIVAL); err != nil {
			return fmt.Errorf("app.taxiCall.DriverToArrival: invalid state change: %w", err)
		}

		if err := d.repository.taxiCallRequest.Update(ctx, i, taxiCallRequest); err != nil {
			return fmt.Errorf("app.taxiCall.DriverToArrival: error while update taxi call request: %w", err)
		}

		processMessage := command.NewTaxiCallProgressCommand(
			taxiCallRequest.Id,
			taxiCallRequest.CurrentState,
			taxiCallRequest.UpdateTime,
			taxiCallRequest.UpdateTime,
		)

		if err := d.repository.event.BatchCreate(ctx, i, []entity.Event{processMessage}); err != nil {
			return fmt.Errorf("app.taxiCall.DriverToArrival: error while create event: %w", err)
		}

		analytics.WriteAnalyticsLog(ctx, requestTime, analytics.LogType_DriverTaxiToArrival, analytics.DriverTaxiToArrivalPayload{
			DriverId:                  taxiCallRequest.DriverId.String,
			RequestUserId:             taxiCallRequest.UserId,
			TaxiCallRequestId:         taxiCallRequest.Id,
			DriverLocation:            driverLocation.Location,
			TaxiCallRequestCreateTime: taxiCallRequest.CreateTime,
			AcceptTime:                taxiCallRequestAcceptTime,
		})

		return nil
	})
}

func (t taxicallApp) DoneTaxiCallRequest(ctx context.Context, driverId string, req request.DoneTaxiCallRequest) error {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	var taxiCallRequest entity.TaxiCallRequest
	var err error

	err = t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		taxiCallRequest, err = t.repository.taxiCallRequest.GetById(ctx, i, req.TaxiCallRequestId)
		if err != nil {
			return fmt.Errorf("app.taxiCall.DoneTaxiCallRequest: error while get taxi call request: %w", err)
		}

		if taxiCallRequest.CurrentState.Complete() {
			// TODO (taeykeom) change error code
			return fmt.Errorf("app.taxiCall.DoneTaxiCallRequest: already completed call request: %w", value.ErrAlreadyExpiredCallRequest)
		}

		if taxiCallRequest.DriverId.String != driverId {
			return fmt.Errorf("app.taxiCall.DoneTaxiCallRequest: forbidden access: %w", value.ErrUnAuthorized)
		}

		taxiCallRequestToArrivalTime := taxiCallRequest.UpdateTime

		if err := taxiCallRequest.UpdateState(requestTime, enum.TaxiCallState_DONE); err != nil {
			return fmt.Errorf("app.taxiCall.DoneTaxiCallRequest: invalid state change:%w", err)
		}
		taxiCallRequest.BasePrice = req.BasePrice
		taxiCallRequest.TollFee = req.TollFee
		taxiCallRequest.UpdateTime = requestTime

		if err := t.repository.taxiCallRequest.Update(ctx, i, taxiCallRequest); err != nil {
			return fmt.Errorf("app.taxiCall.DoneTaxiCallRequest: error while update taxi call request :%w", err)
		}

		driverTaxiCallContext, err := t.repository.taxiCallRequest.GetDriverTaxiCallContext(ctx, i, driverId)
		if err != nil {
			return fmt.Errorf("app.taxiCall.RejectTaxiCallRequest: error while get taxi call context:%w", err)
		}

		driverTaxiCallContext.CanReceive = true
		if err := t.repository.taxiCallRequest.UpsertDriverTaxiCallContext(ctx, i, driverTaxiCallContext); err != nil {
			return fmt.Errorf("app.taxiCall.DoneTaxiCallRequest: error while upsert taxi call context: %w", err)
		}

		processMessage := command.NewTaxiCallProgressCommand(
			taxiCallRequest.Id,
			taxiCallRequest.CurrentState,
			taxiCallRequest.UpdateTime,
			taxiCallRequest.UpdateTime,
		)

		if err := t.repository.event.BatchCreate(ctx, i, []entity.Event{processMessage}); err != nil {
			return fmt.Errorf("app.taxiCall.DoneTaxiCallRequest: error while create taxi call process event: %w", err)
		}

		analytics.WriteAnalyticsLog(ctx, requestTime, analytics.LogType_DriverTaxiDone, analytics.DriverTaxiDonePaylod{
			DriverId:                  taxiCallRequest.DriverId.String,
			RequestUserId:             taxiCallRequest.UserId,
			TaxiCallRequestId:         taxiCallRequest.Id,
			BasePrice:                 req.BasePrice,
			RequestBasePrice:          taxiCallRequest.RequestBasePrice,
			AdditionalPrice:           taxiCallRequest.AdditionalPrice,
			DriverLocation:            driverTaxiCallContext.Location,
			TaxiCallRequestCreateTime: taxiCallRequest.CreateTime,
			ToArrivalTime:             taxiCallRequestToArrivalTime,
		})

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
