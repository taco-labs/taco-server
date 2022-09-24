package response

import (
	"time"

	"github.com/taco-labs/taco/go/domain/value"
)

type TaxiCallRequestCard struct {
	PaymentId  string `json:"paymentId"`
	Company    string `json:"company"`
	CardNumber string `json:"cardNumber"`
}

type TaxiCallRequestResponse struct {
	Id                        string                           `json:"id"`
	UserId                    string                           `json:"userId"`
	DriverId                  string                           `json:"driverId"`
	Departure                 value.Location                   `json:"departure"`
	Arrival                   value.Location                   `json:"arrival"`
	Payment                   TaxiCallRequestCard              `json:"payment"`
	RequestBasePrice          int                              `json:"requestBasePrice"`
	RequestMinAdditionalPrice int                              `json:"requestMinAdditionalPrice"`
	RequestMaxAdditionalPrice int                              `json:"requestMaxAdditionalPrice"`
	BasePrice                 int                              `json:"basePrice"`
	AdditionalPrice           int                              `json:"additionalPrice"`
	CallHistory               []TaxiCallRequestHistoryResponse `json:"history"`
	CreateTime                time.Time                        `json:"createTime"`
	UpdateTime                time.Time                        `json:"updateTime"`
}

type TaxiCallRequestHistoryResponse struct {
	Id                string    `json:"id"`
	TaxiCallRequestId string    `json:"taxiCallRequestId"`
	TaxiCallState     string    `json:"taxiCallState"`
	CreateTime        time.Time `json:"createTime"`
}

type TaxiCallRequestPageResponse struct {
	PageToken string                    `json:"pageToken"`
	Data      []TaxiCallRequestResponse `json:"data"`
}
