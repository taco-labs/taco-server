package service

import (
	"context"
	"fmt"

	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/utils/slices"
)

type ServiceRegionChecker interface {
	ListServiceRegion(context.Context) ([]string, error)
	CheckAvailableServiceRegion(context.Context, string) (bool, error)
}

type staticServiceRegionChecker struct {
	availableServiceRegions    []string
	availableServiceRegionMaps map[string]struct{}
}

func (s staticServiceRegionChecker) ListServiceRegion(ctx context.Context) ([]string, error) {
	return s.availableServiceRegions, nil
}

func (s staticServiceRegionChecker) CheckAvailableServiceRegion(ctx context.Context, serviceRegion string) (bool, error) {
	_, ok := s.availableServiceRegionMaps[serviceRegion]

	return ok, nil
}

func NewStaticServiceRegionChecker(availableServiceRegions []string) (*staticServiceRegionChecker, error) {
	if availableServiceRegions == nil {
		availableServiceRegions = []string{}
	}

	for _, availableServiceRegion := range availableServiceRegions {
		if _, ok := value.SupportedServiceRegionMap[availableServiceRegion]; !ok {
			return nil, fmt.Errorf("invalid service region name: %s", availableServiceRegion)
		}
	}

	availableServiceRegionMaps := slices.ToMapWithValue(availableServiceRegions, func(i string) (string, struct{}) {
		return i, struct{}{}
	})

	return &staticServiceRegionChecker{
		availableServiceRegions:    availableServiceRegions,
		availableServiceRegionMaps: availableServiceRegionMaps,
	}, nil
}
