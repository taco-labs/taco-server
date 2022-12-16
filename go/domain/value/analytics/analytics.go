package analytics

type EventType string

const (
	EventType_UserSignup EventType = "User_Signup"

	EventType_UserTaxiCallRequest EventType = "User_TaxiCallRequest"

	EventType_UserCancelTaxiCallRequest EventType = "User_CancelTaxiCallRequest"

	EventType_UserTaxiCallRequestFailed EventType = "User_taxiCallRequestFailed"

	EventType_UserReferralPointReceived EventType = "User_ReferralPointReceived"

	EventType_DriverSignup EventType = "Driver_Signup"

	EventType_DriverOnDuty EventType = "Driver_OnDuty"

	EventType_DriverLocation EventType = "Driver_Location"

	EventType_DriverTaxiCallTicketDistribution EventType = "Driver_TaxiCallTicketDistribution"

	EventType_DriverTaxiCallTicketAccept EventType = "Driver_TaxiCallTicketAccept"

	EventType_DriverTaxiCallTicketReject EventType = "Driver_TaxiCallTicketReject"

	EventType_DriverTaxiCallCancel EventType = "Driver_TaxiCallCancel"

	EventType_DriverTaxiToArrival EventType = "Driver_TaxiToArrival"

	EventType_DriverTaxiDone EventType = "Driver_TaxiDone"
)

type AnalyticsEvent interface {
	EventType() EventType
}
