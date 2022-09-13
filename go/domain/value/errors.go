package value

import (
	"errors"
	"fmt"
)

var (
	ErrUnAuthenticated = errors.New("unauthenticated")

	ErrSessionExpired = errors.New("session expired")

	ErrDBInternal = errors.New("db internal error")

	ErrNotFound = errors.New("not found")

	ErrUserNotFound = fmt.Errorf("%v: user", ErrNotFound)

	ErrDriverNotFound = fmt.Errorf("%v: driver", ErrNotFound)

	ErrExternal = errors.New("external error")

	ErrUnAuthorized = errors.New("unauthorized")
)
