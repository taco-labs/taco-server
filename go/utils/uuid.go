package utils

import (
	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

func MustNewUUID() string {
	ulid := ulid.Make()
	return uuid.Must(uuid.FromBytes(ulid.Bytes())).String()
}
