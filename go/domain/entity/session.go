package entity

import (
	"time"

	"github.com/uptrace/bun"
)

type UserSession struct {
	bun.BaseModel `bun:"table:user_session"`

	Id         string    `bun:"id,pk"`
	UserId     string    `bun:"user_id"`
	ExpireTime time.Time `bun:"expire_time"`
}

type DriverSession struct {
	bun.BaseModel `bun:"table:driver_session"`

	Id         string    `bun:"id,pk"`
	DriverId   string    `bun:"driver_id"`
	ExpireTime time.Time `bun:"expire_time"`
}
