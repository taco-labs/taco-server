package user

import (
	"context"
	"fmt"

	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/value"
)

func (u userApp) SearchLocation(ctx context.Context, req request.SearchLocationRequest) ([]value.LocationSummary, error) {
	fmt.Printf("Test: %++v\n", req)
	point := value.Point{
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	}
	resp, err := u.service.location.SearchLocation(ctx, point, req.Keyword)

	if err != nil {
		return []value.LocationSummary{}, fmt.Errorf("app.user.SearchLocation: error from search location: %w", err)
	}

	return resp, nil
}
