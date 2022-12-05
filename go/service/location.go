package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/utils/slices"
)

const (
	SentinelPageToken = -1
)

type LocationService interface {
	SearchLocation(context.Context, value.Point, string, int) ([]value.LocationSummary, int, error)
	GetAddress(context.Context, value.Point) (value.Address, error)
}

type kakaoLocationService struct {
	client *resty.Client
}

type kakaoLocationSearchDocuments struct {
	PlaceName   string `json:"place_name"`
	AddressName string `json:"address_name"`
	X           string `json:"x"`
	Y           string `json:"y"`
}

func toLocationSummary(k kakaoLocationSearchDocuments) value.LocationSummary {
	longitude, _ := strconv.ParseFloat(k.X, 64)
	latitude, _ := strconv.ParseFloat(k.Y, 64)

	return value.LocationSummary{
		Point: value.Point{
			Longitude: longitude,
			Latitude:  latitude,
		},
		AddressSummary: value.AddressSummary{
			AddressName: k.AddressName,
			PlaceName:   k.PlaceName,
		},
	}
}

type kakaoLocationSearchResponse struct {
	Documents []kakaoLocationSearchDocuments `json:"documents"`
	Meta      struct {
		IsEnd bool `json:"is_end"`
	} `json:"meta"`
}

func (k kakaoLocationService) SearchLocation(ctx context.Context, point value.Point, keyword string, pageToken int) ([]value.LocationSummary, int, error) {
	if pageToken == SentinelPageToken {
		return []value.LocationSummary{}, SentinelPageToken, nil
	}

	resp, err := k.client.R().
		SetQueryParam("query", keyword).
		SetQueryParam("x", fmt.Sprint(point.Longitude)).
		SetQueryParam("y", fmt.Sprint(point.Latitude)).
		SetQueryParam("page", fmt.Sprint(pageToken+1)).
		SetResult(&kakaoLocationSearchResponse{}).
		Get("v2/local/search/keyword.json")

	if err != nil {
		return []value.LocationSummary{}, pageToken, fmt.Errorf("%w: erorr from kakao map search: %v", value.ErrExternal, err)
	}

	searchResp := resp.Result().(*kakaoLocationSearchResponse)

	documents := searchResp.Documents

	var newPageToken int
	if searchResp.Meta.IsEnd {
		newPageToken = SentinelPageToken
	} else {
		newPageToken = pageToken + 1
	}

	return slices.Map(documents, toLocationSummary), newPageToken, nil
}

type kakaoAddress struct {
	AddressName   string `json:"address_name"`
	RegionDepth1  string `json:"region_1depth_name"`
	RegionDepth2  string `json:"region_2depth_name"`
	RegionDepth3  string `json:"region_3depth_name"`
	MainAddressNo string `json:"main_address_no"`
	SubAddressNo  string `json:"sub_address_no"`
}

type kakaoAddressResponse struct {
	Documents []struct {
		Address kakaoAddress `json:"address"`
	} `json:"documents"`
}

func (k kakaoLocationService) GetAddress(ctx context.Context, point value.Point) (value.Address, error) {
	resp, err := k.client.R().
		SetQueryParam("x", fmt.Sprint(point.Longitude)).
		SetQueryParam("y", fmt.Sprint(point.Latitude)).
		SetResult(&kakaoAddressResponse{}).
		Get("v2/local/geo/coord2address.json")

	if err != nil {
		return value.Address{}, fmt.Errorf("%w: error from kakao coord2regioncode: %v", value.ErrExternal, err)
	}

	addressResp := resp.Result().(*kakaoAddressResponse).Documents[0].Address

	address := value.Address{
		AddressName:   addressResp.AddressName,
		RegionDepth1:  addressResp.RegionDepth1,
		RegionDepth2:  addressResp.RegionDepth2,
		RegionDepth3:  addressResp.RegionDepth3,
		MainAddressNo: addressResp.MainAddressNo,
		SubAddressNo:  addressResp.SubAddressNo,
		ServiceRegion: value.GetServiceRegion(addressResp.AddressName),
	}

	return address, nil
}

func NewKakaoLocationService(endpoint string, apiKey string) *kakaoLocationService {
	client := resty.New().
		SetBaseURL(endpoint).
		SetAuthScheme("KakaoAK").
		SetAuthToken(apiKey)

	return &kakaoLocationService{
		client: client,
	}
}
