package request

type ListDriverTaxiCallContextInRadiusRequest struct {
	Latitude      float64 `query:"latitude"`
	Longitude     float64 `query:"longitude"`
	Radius        int     `query:"radius"`
	ServiceRegion string  `query:"serviceRegion"`
}
