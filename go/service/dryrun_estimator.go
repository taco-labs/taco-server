package service

import (
	"context"
	"time"

	"github.com/taco-labs/taco/go/domain/value"
)

type DryRunEstimator interface {
	EstimateAdditionalPriceRange(ctx context.Context, requestTime time.Time, departure value.Location, arrival value.Location, route value.Route) (int, int, error)
}

type mockDryRunEstimator struct{}

func (m mockDryRunEstimator) EstimateAdditionalPriceRange(ctx context.Context, requestTime time.Time, departure value.Location, arrival value.Location, route value.Route) (int, int, error) {
	return 3000, 10000, nil
}

func NewMockDryRunEstimator() *mockDryRunEstimator {
	return &mockDryRunEstimator{}
}

type hourlySlicedStaticEstimator struct {
	minPriceByHours [24]int
	maxPriceByHours [24]int
}

func (h hourlySlicedStaticEstimator) EstimateAdditionalPriceRange(ctx context.Context, requestTime time.Time, departure value.Location, arrival value.Location, route value.Route) (int, int, error) {
	hourInKst := requestTime.In(value.Timezone_Kst).Hour()

	return h.minPriceByHours[hourInKst], h.maxPriceByHours[hourInKst], nil
}

func NewHourlySlicesStaticEstimator(minPriceByHours, maxPriceByHours [24]int) *hourlySlicedStaticEstimator {
	return &hourlySlicedStaticEstimator{
		minPriceByHours: minPriceByHours,
		maxPriceByHours: maxPriceByHours,
	}
}
