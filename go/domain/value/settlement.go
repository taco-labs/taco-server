package value

type SettlementAccount struct {
	BankCode          string
	AccountNumber     string
	AccountHolderName string
	BankTransactionId string
}

type SettlementTransferRequest struct {
	DriverId          string
	TransferKey       string
	BankTransactionId string
	Amount            int
	Message           string // TODO (taekyeom) 10자 이내이여야 함
}

type SettlementTransfer struct {
	ExecutionKey string
}
