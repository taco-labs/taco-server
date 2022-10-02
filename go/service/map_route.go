package service

import (
	"context"

	"github.com/taco-labs/taco/go/domain/value"
)

type MapRouteService interface {
	GetRoute(context.Context, value.Location, value.Location) (value.Route, error)
}

type mockRouteService struct{}

func (m mockRouteService) GetRoute(ctx context.Context, departure value.Location, arrival value.Location) (value.Route, error) {
	return value.Route{ETA: 0, Price: 10000}, nil
}

func NewMockRouteService() mockRouteService {
	return mockRouteService{}
}
