package driver

import (
	"context"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/utils"
)

func (d driverApp) ListTaxiCallRequest(ctx context.Context, req request.ListDriverTaxiCallRequest) ([]entity.TaxiCallRequest, string, error) {
	return d.service.taxiCall.ListDriverTaxiCallRequest(ctx, req)
}

func (d driverApp) GetLatestTaxiCallRequest(ctx context.Context, driverId string) (entity.TaxiCallRequest, error) {
	return d.service.taxiCall.LatestDriverTaxiCallRequest(ctx, driverId)
}

// TODO (taekyeom) Remove it later!!
func (d driverApp) ForceAcceptTaxiCallRequest(ctx context.Context, driverId, callRequestId string) error {
	return d.service.taxiCall.ForceAcceptTaxiCallRequest(ctx, driverId, callRequestId)
}

func (d driverApp) DriverToArrival(ctx context.Context, callRequestId string) error {
	driverId := utils.GetDriverId(ctx)
	return d.service.taxiCall.DriverToArrival(ctx, driverId, callRequestId)
}

func (d driverApp) AcceptTaxiCallRequest(ctx context.Context, ticketId string) error {
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
