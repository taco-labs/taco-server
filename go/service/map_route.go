package service

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/taco-labs/taco/go/domain/value"
)

type MapRouteService interface {
	GetRoute(context.Context, value.Point, value.Point) (value.Route, error)
}

type mockRouteService struct{}

func (m mockRouteService) GetRoute(ctx context.Context, departure value.Point, arrival value.Point) (value.Route, error) {
	return value.Route{ETA: 0, Price: 10000}, nil
}

func NewMockRouteService() mockRouteService {
	return mockRouteService{}
}

type naverMapsRouteService struct {
	client *resty.Client
}

type naverMapsRouteUnit struct {
	Summary struct {
		Duration int `json:"duration"`
		TaxiFare int `json:"taxiFare"`
		Distance int `json:"distance"`
	} `json:"summary"`
}

type naverMapsRouteResponse struct {
	Code    int                              `json:"code"`
	Message string                           `json:"message"`
	Route   map[string][]*naverMapsRouteUnit `json:"route"`
}

func (m naverMapsRouteService) GetRoute(ctx context.Context, departure value.Point, arrival value.Point) (value.Route, error) {
	resp, err := m.client.R().
		SetQueryParam("start", departure.Format()).
		SetQueryParam("goal", arrival.Format()).
		SetQueryParam("option", "traoptimal").
		SetResult(&naverMapsRouteResponse{}).
		Get("map-direction/v1/driving")

	if err != nil {
		return value.Route{}, fmt.Errorf("%w: error from naver maps api: %v", value.ErrExternal, err)
	}

	naverMapsRouteResp := resp.Result().(*naverMapsRouteResponse)

	if naverMapsRouteResp.Code != 0 {
		return value.Route{}, fmt.Errorf("%w: invalid route", value.ErrInvalidRoute)
	}

	// TODO (taekyeom) handle waypoints
	routeSummary := naverMapsRouteResp.Route["traoptimal"][0].Summary

	return value.Route{
		ETA:      time.Millisecond * time.Duration(routeSummary.Duration),
		Price:    routeSummary.TaxiFare,
		Distance: routeSummary.Distance,
	}, nil
}

func NewNaverMapsRouteService(endpoint string, clientKey string, clientSecret string) naverMapsRouteService {
	client := resty.New().
		SetBaseURL(endpoint).
		SetHeader("X-NCP-APIGW-API-KEY-ID", clientKey).
		SetHeader("X-NCP-APIGW-API-KEY", clientSecret)

	return naverMapsRouteService{
		client: client,
	}
}
