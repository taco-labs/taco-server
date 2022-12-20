package value

import "github.com/google/uuid"

var (
	MockDriverId = uuid.Nil.String()
	MockUserId   = uuid.UUID([16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}).String()
)
