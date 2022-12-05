package enum

type SettlementTransferProcessState string

const (
	SettlementTransferProcessState_Received SettlementTransferProcessState = "TRANSFER_REQUEST_RECEIVED"

	SettlementTransferProcessState_REQUESTED SettlementTransferProcessState = "TRANSFER_REQUESTED"

	SettlementTransferProcessState_EXECUTED SettlementTransferProcessState = "TRANSFER_EXECUTED"

	SettlementTransferProcessState_SUCCEEDED SettlementTransferProcessState = "TRANSFER_SUCCEEDED"

	SettlementTransferProcessState_FAILED SettlementTransferProcessState = "TRANSFER_FAILED"
)
