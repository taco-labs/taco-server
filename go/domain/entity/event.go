package entity

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/taco-labs/taco/go/domain/value"
	"github.com/uptrace/bun"
)

const (
	MetaDataKey_EventUri = "event_uri"
)

type Event struct {
	bun.BaseModel `bun:"table:event"`

	MessageId    string          `bun:"message_id,pk"`
	EventUri     string          `bun:"event_uri"`
	DelaySeconds int64           `bun:"delay_seconds"`
	Payload      json.RawMessage `bun:"payload,type:jsonb"`
	CreateTime   time.Time       `bun:"create_time,nullzero,notnull,default:current_timestamp"`
	AttemtCount  int             `bun:"-"`
	ackFn        func() error
}

func (e *Event) SetAck(fn func() error) {
	e.ackFn = fn
}

func (e Event) Ack() error {
	if e.ackFn == nil {
		return fmt.Errorf("%w: nil ack function", value.ErrInvalidOperation)
	}

	return e.ackFn()
}
