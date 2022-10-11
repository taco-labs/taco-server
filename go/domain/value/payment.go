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

type PaymentSummary struct {
	PaymentId  string `json:"paymentId"`
	Company    string `json:"company"`
	CardNumber string `json:"cardNumber"`
}

type CardPaymentInfo struct {
	CustomerKey         string
	CardCompany         string
	CardNumber          string
	CardExpirationYear  string
	CardExpirationMonth string
	BillingKey          string
}
