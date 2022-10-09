package value

import (
	"fmt"
	"time"
)

type Address struct {
	AddressName   string `json:"addressName"`
	RegionDepth1  string `json:"regionDepth1"`
	RegionDepth2  string `json:"regionDepth2"`
	RegionDepth3  string `json:"regionDepth3"`
	MainAddressNo string `json:"mainAddressNo"`
	SubAddressNo  string `json:"subAddressNo"`
}

func (r Address) AvailableRegion() bool {
	return r.RegionDepth1 == "서울" && (r.RegionDepth2 == "서초구" || r.RegionDepth2 == "강남구")
}

type AddressSummary struct {
	PlaceName   string `json:"placeName"`
	AddressName string `json:"addressName"`
}

type Point struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func (p Point) Format() string {
	return fmt.Sprintf("%f,%f", p.Longitude, p.Latitude)
}

type Location struct {
	Address Address `json:"address"`
	Point   Point   `json:"point"`
}

type LocationSummary struct {
	AddressSummary AddressSummary `json:"addressSummary"`
	Point          Point          `json:"point"`
}

type Route struct {
	ETA   time.Duration `json:"eta"`
	Price int           `json:"price"`
}
