package value

import (
	"fmt"
)

type ErrCode string

const (
	ERR_UNAUTHENTICATED   ErrCode = "ERR_UNAUTHENTICATED"
	ERR_UNAUTHORIZED      ErrCode = "ERR_UNAUTHORIZED"
	ERR_SESSION_EXPIRED   ErrCode = "ERR_SESSION_EXPIRED"
	ERR_DB_INTERNAL       ErrCode = "ERR_DB_INTERNAL"
	ERR_NOTFOUND          ErrCode = "ERR_NOT_FOUND"
	ERR_EXTERNAL          ErrCode = "ERR_EXTERNAL"
	ERR_ALREADY_EXISTS    ErrCode = "ERR_ALREADY_EXISTS"
	ERR_INVALID           ErrCode = "ERR_INVALID"
	ERR_INTERNAL          ErrCode = "ERR_INTERNAL"
	ERR_NEED_CONFIRMATION ErrCode = "ERR_NEED_CONFIRMATION"
	ERR_UNSUPPORTED       ErrCode = "ERR_UNSUPPORTED"
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

	ErrInvalidTaxiCallStateTransition = TacoError{ERR_INVALID, "invalid taxi call state change"}

	ErrConfirmationNeededStateTransition = TacoError{ERR_NEED_CONFIRMATION, "need confirmation"}

	ErrUnsupportedRegion = TacoError{ERR_UNSUPPORTED, "unsupported region"}
)
