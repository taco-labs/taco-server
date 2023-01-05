package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/utils/slices"
)

const (
	searchLocationPath = "maps/v3.0/appkeys/{appKey}/searches"
	getAddressPath     = "/maps/v3.0/appkeys/{appKey}/addresses"
	getRoutePath       = "/maps/v3.0/appkeys/{appKey}/route-normal"

	PageCountDefault = 20
)

type MapService interface {
	SearchLocation(context.Context, value.Point, string, int, int) ([]value.LocationSummary, int, error)
	GetAddress(context.Context, value.Point) (value.Address, error)
	GetRoute(context.Context, value.Point, value.Point) (value.Route, error)
}

type mockMapService struct{}

func (m mockMapService) SearchLocation(context.Context, value.Point, string, int, int) ([]value.LocationSummary, int, error) {
	return []value.LocationSummary{}, 0, nil
}

func (m mockMapService) GetAddress(context.Context, value.Point) (value.Address, error) {
	return value.NewAddress("", "서울", "", "", "", "", ""), nil
}

func (m mockMapService) GetRoute(context.Context, value.Point, value.Point) (value.Route, error) {
	return value.Route{}, nil
}

func NewMockMapService() *mockMapService {
	return &mockMapService{}
}

type nhnMapsService struct {
	client *resty.Client
}

func (n nhnMapsService) SearchLocation(ctx context.Context, point value.Point, keyword string, pageToken int, pageCount int) ([]value.LocationSummary, int, error) {
	if pageCount == 0 {
		pageCount = PageCountDefault
	}

	resp, err := n.client.R().
		SetQueryParams(map[string]string{
			"query":         keyword,
			"coordtype":     "1",
			"startposition": fmt.Sprint(pageToken),
			"reqcount":      fmt.Sprint(pageCount),
			"spopt":         "0",
			"depth":         "0",
			"x1":            fmt.Sprint(point.Longitude),
			"y1":            fmt.Sprint(point.Latitude),
			"sortopt":       "3",
		}).
		SetResult(&nhnMapsSearchLocationResponse{}).
		Get(searchLocationPath)

	if err != nil {
		return []value.LocationSummary{}, 0, fmt.Errorf("%w: error from nhn maps api: %v", value.ErrExternal, err)
	}

	nhnMapsSearchResp := resp.Result().(*nhnMapsSearchLocationResponse)

	if err := nhnMapsSearchResp.Err(); err != nil {
		return []value.LocationSummary{}, 0, err
	}

	var nextPageToken int
	if nhnMapsSearchResp.End() {
		nextPageToken = pageToken
	} else {
		nextPageToken = pageToken + nhnMapsSearchResp.Search.Count
	}

	return slices.Map(nhnMapsSearchResp.Search.Locations, nhnMapsLocationResponseToLocationSummary), nextPageToken, nil
}

func (n nhnMapsService) GetAddress(ctx context.Context, point value.Point) (value.Address, error) {
	resp, err := n.client.R().
		SetQueryParams(map[string]string{
			"posX":      fmt.Sprint(point.Longitude),
			"posY":      fmt.Sprint(point.Latitude),
			"coordtype": "1",
		}).
		SetResult(&nhnMapsServiceGetAddressResponse{}).
		Get(getAddressPath)

	if err != nil {
		return value.Address{}, fmt.Errorf("%w: error from nhn maps api: %v", value.ErrExternal, err)
	}

	nhnMapAddressResp := resp.Result().(*nhnMapsServiceGetAddressResponse)
	if err := nhnMapAddressResp.Err(); err != nil {
		return value.Address{}, err
	}

	return nhnMapsServiceGetAddressResponseToAddress(nhnMapAddressResp), nil
}

func (n nhnMapsService) GetRoute(ctx context.Context, departure value.Point, arrival value.Point) (value.Route, error) {
	resp, err := n.client.R().
		SetQueryParams(map[string]string{
			"coordType":   "WGS84",
			"startX":      fmt.Sprint(departure.Longitude),
			"startY":      fmt.Sprint(departure.Latitude),
			"endX":        fmt.Sprint(arrival.Longitude),
			"endY":        fmt.Sprint(arrival.Latitude),
			"option":      "recommendation",
			"useTaxifare": "1",
			"usageType":   "1",
		}).
		SetResult(&nhmMapsServiceGetRouteResponse{}).
		Get(getRoutePath)

	if err != nil {
		return value.Route{}, fmt.Errorf("%w: error from nhn maps api: %v", value.ErrExternal, err)
	}

	nhnMapRouteResp := resp.Result().(*nhmMapsServiceGetRouteResponse)
	if err := nhnMapRouteResp.Err(); err != nil {
		return value.Route{}, err
	}

	path := slices.FlatMap(nhnMapRouteResp.Route.Data.Path, func(i nhnMapsServiceRoutePath) []value.Point {
		return slices.Map(i.Coords, func(i nhnMapsRouteCoord) value.Point {
			return value.Point{
				Longitude: i.Longitude,
				Latitude:  i.Latitude,
			}
		})
	})

	return value.Route{
		ETA:      time.Second * time.Duration(nhnMapRouteResp.Route.Data.SpendTime),
		Price:    nhnMapRouteResp.Route.Data.TaxiFare,
		Distance: nhnMapRouteResp.Route.Data.Distnace,
		Path:     path,
	}, nil
}

func NewNhnMapsService(endpoint string, apiKey string) *nhnMapsService {
	client := resty.New().
		SetBaseURL(endpoint).
		SetPathParam("appKey", apiKey)

	return &nhnMapsService{
		client: client,
	}
}

type nhnMapsResponseHeader struct {
	Header struct {
		Success       bool   `json:"isSuccessful"`
		ResultCode    int    `json:"resultCode"`
		ResultMessage string `json:"resultMessage"`
	} `json:"header"`
}

func (n nhnMapsResponseHeader) Err() error {
	if n.Header.Success {
		return nil
	}

	return fmt.Errorf("%w: error from response: %v (code: %v)", value.ErrExternal, n.Header.ResultMessage, n.Header.ResultCode)
}

type nhnMapsLocationResponse struct {
	Longitude string `json:"dpx"`
	Latitude  string `json:"dpy"`
	PlaceName string `json:"name1"`
	Address   string `json:"address"`
}

func nhnMapsLocationResponseToLocationSummary(n nhnMapsLocationResponse) value.LocationSummary {
	longitude, _ := strconv.ParseFloat(n.Longitude, 64)
	latitude, _ := strconv.ParseFloat(n.Latitude, 64)
	return value.LocationSummary{
		Point: value.Point{
			Longitude: longitude,
			Latitude:  latitude,
		},
		AddressSummary: value.AddressSummary{
			AddressName: n.Address,
			PlaceName:   n.PlaceName,
		},
	}
}

type nhnMapsSearchLocationResponse struct {
	nhnMapsResponseHeader
	Search struct {
		Count     int                       `json:"count"`
		Locations []nhnMapsLocationResponse `json:"poi"`
	} `json:"search"`
}

func (n nhnMapsSearchLocationResponse) End() bool {
	return n.Header.ResultCode == 100
}

type nhnMapsRouteCoord struct {
	Longitude float64 `json:"x"`
	Latitude  float64 `json:"y"`
}

type nhnMapsServiceRoutePath struct {
	Coords []nhnMapsRouteCoord `json:"coords"`
}

type nhmMapsServiceGetRouteResponse struct {
	nhnMapsResponseHeader
	Route struct {
		Data struct {
			SpendTime int                       `json:"spend_time"`
			Distnace  int                       `json:"distance"`
			TollFee   int                       `json:"toll_fee"`
			TaxiFare  int                       `json:"taxiFare"`
			Path      []nhnMapsServiceRoutePath `json:"paths"`
		} `json:"data"`
	} `json:"route"`
}

type nhnMapsServiceGetAddressResponse struct {
	nhnMapsResponseHeader
	Location struct {
		HasAdmAddress bool `json:"hasAdmAddress"`
		Adm           struct {
			BuildingName string `json:"bldname"`
		} `json:"adm"`
		AdministrativeAddress nhnAddress `json:"adm_address"`
		LegalAddress          nhnAddress `json:"legal_address"`
	} `json:"location"`
}

type nhnAddress struct {
	Address      string `json:"cut_address"`
	RegionDepth1 string `json:"address_category1"`
	RegionDepth2 string `json:"address_category2"`
	RegionDepth3 string `json:"address_category3"`
	RegionDepth4 string `json:"address_category4"`
	AddressNo    string `json:"jibun"`
}

func nhnMapsServiceGetAddressResponseToAddress(n *nhnMapsServiceGetAddressResponse) value.Address {
	var addressToTransform nhnAddress
	if n.Location.HasAdmAddress {
		addressToTransform = n.Location.AdministrativeAddress
	} else {
		addressToTransform = n.Location.LegalAddress
	}

	addressName := fmt.Sprintf("%s %s", addressToTransform.Address, addressToTransform.AddressNo)
	regionDepth1 := addressToTransform.RegionDepth1
	regionDepth2 := addressToTransform.RegionDepth2
	var regionDepth3 string
	var mainAddressNo string
	var subAddressNo string

	if addressToTransform.RegionDepth4 != "" {
		regionDepth3 = fmt.Sprintf("%s %s", addressToTransform.RegionDepth3, addressToTransform.RegionDepth4)
	} else {
		regionDepth3 = addressToTransform.RegionDepth3
	}

	addressNumberParts := strings.Split(addressToTransform.AddressNo, "-")
	if len(addressNumberParts) == 2 {
		mainAddressNo = addressNumberParts[0]
		subAddressNo = addressNumberParts[1]
	} else {
		mainAddressNo = addressNumberParts[0]
	}

	return value.NewAddress(
		addressName,
		regionDepth1,
		regionDepth2,
		regionDepth3,
		mainAddressNo,
		subAddressNo,
		n.Location.Adm.BuildingName,
	)
}
