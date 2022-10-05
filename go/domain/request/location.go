package request

type SearchLocationRequest struct {
	Latitude  float64 `query:"latitude"`
	Longitude float64 `query:"longitude"`
	Keyword   string  `query:"keyword"`
}
