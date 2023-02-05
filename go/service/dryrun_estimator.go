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

// TODO (taekyeom) make dsl...
type adhocStaticEstimator struct{}

func (a adhocStaticEstimator) EstimateAdditionalPriceRange(ctx context.Context, requestTime time.Time, departure value.Location, arrival value.Location, route value.Route) (int, int, error) {
	hourInKst := requestTime.In(value.Timezone_Kst).Hour()
	price := route.Price

	handlePriceOver20000 := func(price int) int {
		priceToDiscount := -(price / 10)
		if priceToDiscount > -3000 {
			priceToDiscount = -3000
		}
		if priceToDiscount < -5000 {
			priceToDiscount = -5000
		}

		return priceToDiscount
	}

	if price < 10000 {
		if 2 <= hourInKst && hourInKst < 7 {
			return -2000, 3000, nil
		} else if 7 <= hourInKst && hourInKst < 10 {
			return 0, 5000, nil
		} else if 10 <= hourInKst && hourInKst < 17 {
			return -2000, 5000, nil
		} else if 17 <= hourInKst && hourInKst < 23 {
			return 0, 5000, nil
		} else {
			return 3000, 10000, nil
		}
	} else if price < 20000 {
		if 2 <= hourInKst && hourInKst < 7 {
			return -3000, 3000, nil
		} else if 7 <= hourInKst && hourInKst < 10 {
			return 0, 5000, nil
		} else if 10 <= hourInKst && hourInKst < 17 {
			return -3000, 4000, nil
		} else if 17 <= hourInKst && hourInKst < 23 {
			return 0, 5000, nil
		} else {
			return 3000, 7000, nil
		}
	} else {
		if 2 <= hourInKst && hourInKst < 7 {
			return handlePriceOver20000(price), 3000, nil
		} else if 7 <= hourInKst && hourInKst < 10 {
			return 0, 3000, nil
		} else if 10 <= hourInKst && hourInKst < 17 {
			return handlePriceOver20000(price), 2000, nil
		} else if 17 <= hourInKst && hourInKst < 23 {
			return 0, 4000, nil
		} else {
			return 0, 6000, nil
		}
	}
}

func NewAdhocStaticEstimator() *adhocStaticEstimator {
	return &adhocStaticEstimator{}
}
