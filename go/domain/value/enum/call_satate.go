package enum

type TaxiCallState string

var (
	TaxiCallState_Requested TaxiCallState = "TAXI_CALL_REQUESTED"

	TaxiCallState_DRIVER_TO_DEPARTURE TaxiCallState = "DRIVER_TO_DEPARTURE"

	TaxiCallState_DRIVER_TO_ARRIVAL TaxiCallState = "DRIVER_TO_ARRIVAL"

	TaxiCallState_DONE TaxiCallState = "TAXI_CALL_DONE"

	TaxiCallState_USER_CANCELLED TaxiCallState = "USER_CANCELLED"

	TaxiCallState_DRIVER_CANCELLED TaxiCallState = "DRIVER_CANCELLED"

	TaxiCallState_FAILED TaxiCallState = "TAXI_CALL_FAILED"

	TaxiCallState_SETTLEMENT_DONE TaxiCallState = "DRIVER_SETTLEMENT_DONE"
)
