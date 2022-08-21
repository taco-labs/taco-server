package value

import "errors"

var (
	ErrUnAuthenticated = errors.New("unauthenticated")

	ErrSessionExpired = errors.New("session expired")

	ErrDBInternal = errors.New("db internal error")

	ErrUserNotFound = errors.New("user not found")
)
