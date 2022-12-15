package entity

import (
	"time"

	"github.com/taco-labs/taco/go/domain/value/enum"
	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:\"user\""`

	Id            string      `bun:"id,pk"`
	FirstName     string      `bun:"first_name"`
	LastName      string      `bun:"last_name"`
	BirthDay      string      `bun:"birthday"`
	Phone         string      `bun:"phone"`
	Gender        string      `bun:"gender"`
	AppOs         enum.OsType `bun:"app_os"`
	AppVersion    string      `bun:"app_version"`
	UserUniqueKey string      `bun:"user_unique_key"`
	CreateTime    time.Time   `bun:"create_time"`
	UpdateTime    time.Time   `bun:"update_time"`
	DeleteTime    time.Time   `bun:"delete_time"`

	// TODO (taekyeom) seperate entity & make dto
	UserPoint int `bun:"-"`
}
