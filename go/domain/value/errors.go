package value

import (
	"fmt"
)

type ErrCode string

const (
	ERR_UNAUTHENTICATED      ErrCode = "ERR_UNAUTHENTICATED"
	ERR_UNAUTHORIZED         ErrCode = "ERR_UNAUTHORIZED"
	ERR_SESSION_EXPIRED      ErrCode = "ERR_SESSION_EXPIRED"
	ERR_NOT_YET_ACTIVATED    ErrCode = "ERR_NOT_YET_ACTIVATED"
	ERR_DB_INTERNAL          ErrCode = "ERR_DB_INTERNAL"
	ERR_NOTFOUND             ErrCode = "ERR_NOT_FOUND"
	ERR_EXTERNAL             ErrCode = "ERR_EXTERNAL"
	ERR_EXTERNAL_PAYMENT     ErrCode = "ERR_EXTERNAL_PAYMENT"
	ERR_ALREADY_EXISTS       ErrCode = "ERR_ALREADY_EXISTS"
	ERR_INVALID              ErrCode = "ERR_INVALID"
	ERR_INTERNAL             ErrCode = "ERR_INTERNAL"
	ERR_NEED_CONFIRMATION    ErrCode = "ERR_NEED_CONFIRMATION"
	ERR_UNSUPPORTED          ErrCode = "ERR_UNSUPPORTED"
	ERR_CALL_REQUEST_FAILED  ErrCode = "ERR_CALL_REQUEST_FAILED"
	ERR_CALL_REQUEST_EXPIRED ErrCode = "ERR_CALL_REQUEST_EXPIRED"
	ERR_INVALID_USER_PAYMENT ErrCode = "ERR_INVALID_USER_PAYMENT"

	ERR_PAYMENT_DUPLICATED_ORDER        ErrCode = "ERR_PAYMENT_DUPLICATED_ORDER"
	ERR_PAYMENT_INVALID_CARD_EXPIRATION ErrCode = "ERR_PAYMENT_INVALID_CARD_EXPIRATION"
	ERR_PAYMENT_INVALID_CARD_NUMBER     ErrCode = "ERR_PAYMENT_INVALID_CARD_NUMBER"
	ERR_PAYMENT_INVALID_STOPPED_CARD    ErrCode = "ERR_PAYMENT_INVALID_STOPPED_CARD"
	ERR_PAYMENT_REJECT_ACCOUNT_PAYMENT  ErrCode = "ERR_PAYMENT_REJECT_ACCOUNT_PAYMENT"
)

type TacoError struct {
	ErrCode ErrCode `json:"errCode"`
	Message string  `json:"message"`
}

func (t TacoError) Error() string {
	return fmt.Sprintf("taco error [%s]: %s", t.ErrCode, t.Message)
}

func (t TacoError) Is(target error) bool {
	targetTacoErr, ok := target.(TacoError)
	return ok && targetTacoErr.ErrCode == t.ErrCode
}

var (
	ErrUnsupported = TacoError{ERR_UNSUPPORTED, "unsupported"}

	ErrUnAuthenticated = TacoError{ERR_UNAUTHENTICATED, "unauthenticated"}

	ErrUnAuthorized = TacoError{ERR_UNAUTHORIZED, "unauthorized"}

	ErrSessionExpired = TacoError{ERR_SESSION_EXPIRED, "session expired"}

	ErrDBInternal = TacoError{ERR_DB_INTERNAL, "db internal error"}

	ErrUserNotFound = TacoError{ERR_NOTFOUND, "user not found"}

	ErrDriverNotFound = TacoError{ERR_NOTFOUND, "driver not found"}

	ErrNotFound = TacoError{ERR_NOTFOUND, "not found"}

	ErrExternal = TacoError{ERR_EXTERNAL, "external service error"}

	ErrAlreadyExists = TacoError{ERR_ALREADY_EXISTS, "already exists"}

	ErrInvalidOperation = TacoError{ERR_INVALID, "invalid operation"}

	ErrInvalidRoute = TacoError{ERR_INVALID, "invalid route request"}

	ErrInvalidLocation = TacoError{ERR_INVALID, "invalid location"}

	ErrInvalidTaxiCallStateTransition = TacoError{ERR_INVALID, "invalid taxi call state change"}

	ErrConfirmationNeededStateTransition = TacoError{ERR_NEED_CONFIRMATION, "need confirmation"}

	ErrUnsupportedServiceRegion = TacoError{ERR_UNSUPPORTED, "unsupported region"}

	ErrNotYetActivated = TacoError{ERR_NOT_YET_ACTIVATED, "not yet activated account"}

	ErrInternal = TacoError{ERR_INTERNAL, "internal error"}

	ErrRequestPriceLimitExceed = TacoError{ERR_CALL_REQUEST_FAILED, "additional price limit exceeded"}

	ErrActiveTaxiCallRequestExists = TacoError{ERR_ALREADY_EXISTS, "active taxi call exists"}

	ErrAlreadyExpiredCallRequest = TacoError{ERR_CALL_REQUEST_EXPIRED, "taxi call request expired"}

	ErrInvalidUserPayment = TacoError{ERR_INVALID_USER_PAYMENT, "invalid user payment"}

	ErrExternalPayment = TacoError{ERR_EXTERNAL_PAYMENT, "external payment service error"}

	ErrPaymentDuplicatedOrder = TacoError{ERR_PAYMENT_DUPLICATED_ORDER, "duplicated order"}

	ErrPaymentInvalidCardExpiration = TacoError{ERR_PAYMENT_INVALID_CARD_EXPIRATION, "invalid card expiration"}

	ErrPaymentInvalidCardNumber = TacoError{ERR_PAYMENT_INVALID_CARD_EXPIRATION, "invalid card number"}

	ErrPaymentInvalidStoppedCard = TacoError{ERR_PAYMENT_INVALID_STOPPED_CARD, "stopped card"}

	ErrPaymentRejectAccountPayment = TacoError{ERR_PAYMENT_REJECT_ACCOUNT_PAYMENT, "reject account payment"}
)

func NewTacoError(errCode ErrCode, message string) TacoError {
	return TacoError{
		ErrCode: errCode,
		Message: message,
	}
}
