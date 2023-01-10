package response

import (
	"time"
)

type PaymentSummaryResponse struct {
	PaymentId   string    `json:"paymentId"`
	Company     string    `json:"company"`
	CardNumber  string    `json:"cardNumber"`
	LastUseTime time.Time `json:"lastUseTime"`
}
