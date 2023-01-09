package entity

import (
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/domain/value/enum"
)

func CheckPaymentForTaxiCallRequest(taxiCallRequest TaxiCallRequest, userPayment UserPayment) error {
	if userPayment.Invalid {
		return value.ErrInvalidUserPayment
	}

	switch userPayment.PaymentType {
	case enum.PaymentType_SignupPromition:
		if taxiCallRequest.RequestMaxAdditionalPrice > SignupPromotionMaxAdditionalPrice {
			return value.ErrInvalidPromotionPaymentUsage
		}
	}
	return nil
}
