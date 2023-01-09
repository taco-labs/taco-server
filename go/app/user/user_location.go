package user

import (
	"context"
	"fmt"

	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/value"
)

func (u userApp) GetAddress(ctx context.Context, req request.GetAddressRequest) (value.Address, error) {
	point := value.Point{
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	}

	resp, err := u.service.mapService.GetAddress(ctx, point)
	if err != nil {
		return value.Address{}, fmt.Errorf("app.user.GetAddress: error from get address: %w", err)
	}

	return resp, nil
}

func (u userApp) SearchLocation(ctx context.Context, req request.SearchLocationRequest) ([]value.LocationSummary, int, error) {
	if req.Keyword == "" {
		return []value.LocationSummary{}, 0, nil
	}

	point := value.Point{
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	}
	// TOOD (taekyeom) suppress error as client could not correctly handle error
	resp, pageToken, _ := u.service.mapService.SearchLocation(ctx, point, req.Keyword, req.PageToken, req.PageCount)

	return resp, pageToken, nil
}
