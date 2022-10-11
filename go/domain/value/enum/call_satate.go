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

	TaxiCallState_INVALID TaxiCallState = "INVALID"

	// DO not persist
	TaxiCallState_DRYRUN TaxiCallState = "DRYRUN"
)

func (t TaxiCallState) Active() bool {
	return t == TaxiCallState_Requested ||
		t == TaxiCallState_DRIVER_TO_DEPARTURE ||
		t == TaxiCallState_DRIVER_TO_ARRIVAL
}

func (t TaxiCallState) Requested() bool {
	return t == TaxiCallState_Requested
}

func (t TaxiCallState) InDriving() bool {
	return t == TaxiCallState_DRIVER_TO_DEPARTURE ||
		t == TaxiCallState_DRIVER_TO_ARRIVAL
}

func (t TaxiCallState) Complete() bool {
	return !t.Active()
}

// TODO(taekyeom) handle user / driver cancel
func (t TaxiCallState) TryChangeState(nextState TaxiCallState) bool {
	switch t {
	case TaxiCallState_Requested:
		return nextState == TaxiCallState_DRIVER_TO_DEPARTURE ||
			nextState == TaxiCallState_USER_CANCELLED ||
			nextState == TaxiCallState_FAILED
	case TaxiCallState_DRIVER_TO_DEPARTURE:
		return nextState == TaxiCallState_DRIVER_TO_ARRIVAL ||
			nextState == TaxiCallState_DRIVER_CANCELLED
	case TaxiCallState_DRIVER_TO_ARRIVAL:
		return nextState == TaxiCallState_DONE
	}
	return false
}
