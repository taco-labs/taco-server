package driver

import (
	"context"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/utils"
)

func (d driverApp) ListTaxiCallRequest(ctx context.Context, req request.ListDriverTaxiCallRequest) ([]entity.TaxiCallRequest, string, error) {
	return d.service.taxiCall.ListDriverTaxiCallRequest(ctx, req)
}

func (d driverApp) GetLatestTaxiCallRequest(ctx context.Context, driverId string) (entity.DriverLatestTaxiCallRequest, error) {
	return d.service.taxiCall.LatestDriverTaxiCallRequest(ctx, driverId)
}

// TODO (taekyeom) Remove it later!!
func (d driverApp) ForceAcceptTaxiCallRequest(ctx context.Context, driverId, callRequestId string) (entity.DriverLatestTaxiCallRequest, error) {
	return d.service.taxiCall.ForceAcceptTaxiCallRequest(ctx, driverId, callRequestId)
}

func (d driverApp) DriverToArrival(ctx context.Context, callRequestId string) error {
	driverId := utils.GetDriverId(ctx)
	return d.service.taxiCall.DriverToArrival(ctx, driverId, callRequestId)
}

func (d driverApp) AcceptTaxiCallRequest(ctx context.Context, ticketId string) (entity.DriverLatestTaxiCallRequest, error) {
	driverId := utils.GetDriverId(ctx)
	return d.service.taxiCall.AcceptTaxiCallRequest(ctx, driverId, ticketId)
}

func (d driverApp) RejectTaxiCallRequest(ctx context.Context, ticketId string) error {
	driverId := utils.GetDriverId(ctx)

	return d.service.taxiCall.RejectTaxiCallRequest(ctx, driverId, ticketId)
}

func (d driverApp) DoneTaxiCallRequest(ctx context.Context, req request.DoneTaxiCallRequest) error {
	driverId := utils.GetDriverId(ctx)

	return d.service.taxiCall.DoneTaxiCallRequest(ctx, driverId, req)
}

func (d driverApp) CancelTaxiCallRequest(ctx context.Context, req request.DriverCancelTaxiCallRequest) error {
	driverId := utils.GetDriverId(ctx)
	return d.service.taxiCall.DriverCancelTaxiCallRequest(ctx, driverId, req)
}

func (d driverApp) DriverLatestTaxiCallTicket(ctx context.Context, driverId string) (entity.DriverLatestTaxiCallRequestTicket, error) {
	return d.service.taxiCall.DriverLatestTaxiCallTicket(ctx, driverId)
}

func (d driverApp) AddDriverDenyTaxiCallTag(ctx context.Context, driverId string, tagId int) error {
	return d.service.taxiCall.AddDriverDenyTaxiCallTag(ctx, driverId, tagId)
}

func (d driverApp) DeleteDriverDenyTaxiCallTag(ctx context.Context, driverId string, tagId int) error {
	return d.service.taxiCall.DeleteDriverDenyTaxiCallTag(ctx, driverId, tagId)
}

func (d driverApp) ListDriverDenyTaxiCallTag(ctx context.Context, driverId string) ([]value.Tag, error) {
	return d.service.taxiCall.ListDriverDenyTaxiCallTag(ctx, driverId)
}
