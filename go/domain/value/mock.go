package value

var (
	MockAccountPhoneStart = "01000000000"
	MockAccountPhoneEnd   = "01000010000"

	MockBillingKey = "mock-billing-key"
)

func IsMockPhoneNumber(phone string) bool {
	return phone >= MockAccountPhoneStart && phone < MockAccountPhoneEnd
}

func IsMockPayment(billingKey string) bool {
	return billingKey == MockBillingKey
}
