package value

var (
	MockAccountPhoneStart = "01000000000"
	MockAccountPhoneEnd   = "01000010000"
)

func IsMockPhoneNumber(phone string) bool {
	return phone >= MockAccountPhoneStart && phone < MockAccountPhoneEnd
}
