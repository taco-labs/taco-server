package response

import "github.com/taco-labs/taco/go/domain/entity"

type TaxiCallRequestPageResponse struct {
	PageToken string                   `json:"pageToken"`
	Data      []entity.TaxiCallRequest `json:"data"`
}
