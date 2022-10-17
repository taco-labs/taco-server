package entity

import (
	"time"

	"github.com/uptrace/bun"
)

type PushToken struct {
	bun.BaseModel `bun:"table:push_token"`

	PrincipalId string    `bun:"principal_id,pk"`
	FcmToken    string    `bun:"fcm_token"`
	CreateTime  time.Time `bun:"create_time"`
	UpdateTime  time.Time `bun:"update_time"`
}
