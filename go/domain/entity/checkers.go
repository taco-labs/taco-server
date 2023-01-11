package entity

import (
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/domain/value/enum"
)

func CheckPaymentForTaxiCallRequest(taxiCallRequest *TaxiCallRequest, userPayment UserPayment) error {
	if userPayment.Invalid {
		return value.ErrInvalidUserPayment
	}

	switch userPayment.PaymentType {
	case enum.PaymentType_SignupPromition:
		if taxiCallRequest.RequestMaxAdditionalPrice > SignupPromotionMaxAdditionalPrice {
			return value.ErrInvalidPromotionPaymentUsage
		}
		// TODO (taekyeom) promotion일 경우 min additional price를 최소 5000원으로 설정
		if taxiCallRequest.RequestMinAdditionalPrice < 5000 {
			taxiCallRequest.RequestMinAdditionalPrice = 5000
		}
	}
	return nil
}
