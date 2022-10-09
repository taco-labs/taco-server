package driver

import (
	"context"
	"fmt"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/uptrace/bun"
)

func (d driverApp) ListTaxiCallRequest(ctx context.Context, req request.ListDriverTaxiCallRequest) ([]entity.TaxiCallRequest, string, error) {
	var taxiCallRequests []entity.TaxiCallRequest
	var pageToken string
	var err error

	err = d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		// TODO (settlement?)
		taxiCallRequests, pageToken, err = d.repository.taxiCallRequest.ListByDriverId(ctx, i, req.DriverId, req.PageToken, req.Count)
		if err != nil {
			return fmt.Errorf("app.driver.ListTaxiCallRequest: error while get taxi call requests:%w", err)
		}

		return nil
	})

	if err != nil {
		return []entity.TaxiCallRequest{}, "", err
	}

	return taxiCallRequests, pageToken, nil
}

func (d driverApp) GetLatestTaxiCallRequest(ctx context.Context, driverId string) (entity.TaxiCallRequest, error) {
	var latestTaxiCallRequest entity.TaxiCallRequest
	var err error

	err = d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		latestTaxiCallRequest, err = d.repository.taxiCallRequest.GetLatestByDriverId(ctx, i, driverId)
		if err != nil {
			return fmt.Errorf("app.driver.GetLatestTaxiCallRequest: error while get latest taxi call:\n%w", err)
		}
		return nil
	})

	if err != nil {
		return entity.TaxiCallRequest{}, err
	}

	return latestTaxiCallRequest, nil
}
