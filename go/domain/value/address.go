package value

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkbhex"
)

type number float64

func (n number) MarshalJSON() ([]byte, error) {
	// check given number is integral
	if n == number(int(n)) {
		return []byte(fmt.Sprintf("%.2f", n)), nil
	}
	return []byte(fmt.Sprintf("%f", n)), nil
}

var SupportedServiceRegions = map[string]struct{}{
	"서울":     {},
	"경기 고양":  {},
	"경기 과천":  {},
	"경기 광명":  {},
	"경기 광주":  {},
	"경기 구리":  {},
	"경기 군포":  {},
	"경기 김포":  {},
	"경기 남양주": {},
	"경기 동두천": {},
	"경기 부천":  {},
	"경기 성남":  {},
	"경기 수원":  {},
	"경기 시흥":  {},
	"경기 안산":  {},
	"경기 안성":  {},
	"경기 안양":  {},
	"경기 양주":  {},
	"경기 여주":  {},
	"경기 오산":  {},
	"경기 용인":  {},
	"경기 의왕":  {},
	"경기 의정부": {},
	"경기 이천":  {},
	"경기 파주":  {},
	"경기 평택":  {},
	"경기 포천":  {},
	"경기 하남":  {},
	"경기 화성":  {},
}

type Address struct {
	AddressName   string `json:"addressName"`
	RegionDepth1  string `json:"regionDepth1"`
	RegionDepth2  string `json:"regionDepth2"`
	RegionDepth3  string `json:"regionDepth3"`
	MainAddressNo string `json:"mainAddressNo"`
	SubAddressNo  string `json:"subAddressNo"`
	BuildingName  string `json:"buildingName"`

	// TODO (taekyeom) Address Name 과 Service region 값 사이 불일치가 있을 수 있음.. 일단 별도의 필드로 둔다.
	ServiceRegion string `json:"serviceRegion"`
}

func (r Address) AvailableRegion() bool {
	_, ok := SupportedServiceRegions[r.ServiceRegion]
	return ok
}

func NewAddress(addressName, regionDepth1, regionDepth2, regionDepth3, mainAddressNo, subAddressNo, buildingName string) Address {
	addr := Address{
		AddressName:   addressName,
		RegionDepth1:  regionDepth1,
		RegionDepth2:  regionDepth2,
		RegionDepth3:  regionDepth3,
		MainAddressNo: mainAddressNo,
		SubAddressNo:  subAddressNo,
		BuildingName:  buildingName,
	}
	addr.setServiceRegion()

	return addr
}

func (r *Address) setServiceRegion() {
	for supportedRegion := range SupportedServiceRegions {
		regionParts := strings.Split(supportedRegion, " ")
		if len(regionParts) == 1 && strings.HasPrefix(r.RegionDepth1, regionParts[0]) {
			r.ServiceRegion = supportedRegion
		} else if len(regionParts) == 2 && strings.HasPrefix(r.RegionDepth1, regionParts[0]) && strings.HasPrefix(r.RegionDepth2, regionParts[1]) {
			r.ServiceRegion = supportedRegion
		}
	}
}

type AddressSummary struct {
	PlaceName   string `json:"placeName"`
	AddressName string `json:"addressName"`
}

type Point struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func (p Point) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Latitude  number `json:"latitude"`
		Longitude number `json:"longitude"`
	}{
		Latitude:  number(p.Latitude),
		Longitude: number(p.Longitude),
	})
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
