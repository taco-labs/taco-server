package utils

import "github.com/google/uuid"

func MustNewUUID() string {
	return uuid.Must(uuid.NewUUID()).String()
}
