package request

import (
	"unicode/utf8"

	"github.com/taco-labs/taco/go/domain/value"
)

type CreateTaxiCallRequest struct {
	Dryrun             bool        `json:"dryrun"`
	Departure          value.Point `json:"departure"`
	Arrival            value.Point `json:"arrival"`
	PaymentId          string      `json:"paymentId"`
	MinAdditionalPrice int         `json:"minAdditionalPrice"`
	MaxAdditionalPrice int         `json:"maxAdditionalPrice"`
	TagIds             []int       `json:"tagIds"`
	UserTag            string      `json:"userTag"`
}

// TODO (taekyeom) validation
func (c CreateTaxiCallRequest) Validate() error {
	if utf8.RuneCountInString(c.UserTag) > 10 {
		return value.NewTacoError(value.ERR_INVALID, "요청사항은 10자 이내이어야 합니다.")
	}
	return nil
}

type ListUserTaxiCallRequest struct {
	UserId    string `param:"userId"`
	Count     int    `query:"count"`
	PageToken string `query:"pageToken"`
}

type ListDriverTaxiCallRequest struct {
	DriverId  string `param:"driverId"`
	Count     int    `query:"count"`
	PageToken string `query:"pageToken"`
}

type DoneTaxiCallRequest struct {
	TaxiCallRequestId string `param:"taxiCallRequestId"`
	BasePrice         int    `json:"basePrice"`
	TollFee           int    `json:"tollFee"`
}

type CancelTaxiCallRequest struct {
	TaxiCallRequestId string `param:"taxiCallRequestId"`
	ConfirmCancel     bool   `query:"confirmCancel"`
}
