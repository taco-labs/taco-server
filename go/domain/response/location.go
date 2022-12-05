package response

import "github.com/taco-labs/taco/go/domain/value"

type SearchLocationsResponse struct {
	PageToken int                     `json:"pageToken"`
	Locations []value.LocationSummary `json:"locations"`
}

func SearchLocationsToResopnse(locations []value.LocationSummary, pageToken int) SearchLocationsResponse {
	return SearchLocationsResponse{
		PageToken: pageToken,
		Locations: locations,
	}
}
