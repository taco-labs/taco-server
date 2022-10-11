package value

import "time"

type Route struct {
	ETA      time.Duration `json:"eta"`
	Price    int           `json:"price"`
	Distance int           `json:"distance"`
}
