package request

type UserPaymentRegisterRequest struct {
	Name                   string `bun:"name"`
	CardNumber             string `bun:"card_number"`
	CardExpirationYear     string `bun:"card_expiration_year"`
	CardExpirationMonth    string `bun:"card_expiration_month"`
	CardPassword           string `bun:"card_password"`
	CustomerIdentityNumber string `bun:"customer_identity_number"`
}
