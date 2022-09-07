package value

type Payment struct {
	OrderId   string
	Amount    int
	OrderName string
}

type PaymentResult struct {
	OrderId    string
	PaymentKey string
	Amount     int
	OrderName  string
}
