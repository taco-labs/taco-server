package value

import (
	"fmt"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkbhex"
)

var SupportedServiceRegions = map[string]struct{}{
	"서울": {},
}

type Address struct {
	AddressName   string `json:"addressName"`
	RegionDepth1  string `json:"regionDepth1"`
	RegionDepth2  string `json:"regionDepth2"`
	RegionDepth3  string `json:"regionDepth3"`
	MainAddressNo string `json:"mainAddressNo"`
	SubAddressNo  string `json:"subAddressNo"`
}

func (r Address) AvailableRegion() bool {
	_, ok := SupportedServiceRegions[r.RegionDepth1]
	return ok
}

type AddressSummary struct {
	PlaceName   string `json:"placeName"`
	AddressName string `json:"addressName"`
}

type Point struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func (p Point) ToEwkbHex() (string, error) {
	geomPoint := geom.NewPoint(geom.XY).
		MustSetCoords([]float64{p.Longitude, p.Latitude}).
		SetSRID(SRID_SPHERE)

	ewkbHex, err := ewkbhex.Encode(geomPoint, ewkbhex.NDR)
	if err != nil {
		return "", fmt.Errorf("%w: error while encode location: %v", ErrInternal, err)
	}

	return ewkbHex, err
}

func (p *Point) FromEwkbHex(ewkbHex string) error {
	point, err := ewkbhex.Decode(ewkbHex)
	if err != nil {
		return fmt.Errorf("%w: error while decode location: %v", ErrInternal, err)
	}
	if point.Layout() != geom.XY {
		return fmt.Errorf("%w: invalid location data", ErrInternal)
	}
	coords := point.FlatCoords()
	p.Longitude = coords[0]
	p.Latitude = coords[1]

	return nil
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
