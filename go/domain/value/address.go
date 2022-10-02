package value

import "time"

type RoadAddress struct {
	AddressName  string `json:"addressName"`
	RegionDepth1 string `json:"regionDepth1"`
	RegionDepth2 string `json:"regionDepth2"`
	RegionDepth3 string `json:"regionDepth3"`
	RoadName     string `json:"roadName"`
	BuildingName string `json:"buildingName"`
}

type Location struct {
	RoadAddress RoadAddress `json:"roadAddress"`
	Latitude    float32     `json:"latitude"`
	Longitude   float32     `json:"longitude"`
}

type Route struct {
	ETA   time.Duration `json:"eta"`
	Price int           `json:"price"`
}
