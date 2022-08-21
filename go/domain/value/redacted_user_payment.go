package value

import "time"

type RedactedUserPayment struct {
	Id                             string
	UserId                         string
	Name                           string
	RedactedCardNumber             string
	CardExpirationYear             string
	CardExpirationMonth            string
	RedactedCardPassword           string
	RedactedCustomerIdentityNumber string
	CreateTime                     time.Time
	DeleteTime                     time.Time
}
