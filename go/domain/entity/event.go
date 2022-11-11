package entity

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/taco-labs/taco/go/domain/value"
	"github.com/uptrace/bun"
)

const (
	MetaDataKey_EventUri  = "event_uri"
	MetaDataKey_MessageId = "message_id"
)

type Event struct {
	bun.BaseModel `bun:"table:event"`

	MessageId    string          `bun:"message_id,pk"`
	EventUri     string          `bun:"event_uri"`
	DelaySeconds int32           `bun:"delay_seconds"`
	Payload      json.RawMessage `bun:"payload,type:jsonb"`
	CreateTime   time.Time       `bun:"create_time"`
	RetryCount   int             `bun:"-"`
	ackFn        func() error
	nackFn       func() error
}

func (u *Event) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		u.CreateTime = time.Now()
	}
	return nil
}

func (e *Event) SetAck(fn func() error) {
	e.ackFn = fn
}

func (e *Event) SetNack(fn func() error) {
	e.nackFn = fn
}

func (e Event) Ack() error {
	if e.ackFn == nil {
		return fmt.Errorf("%w: nil ack function", value.ErrInvalidOperation)
	}

	return e.ackFn()
}

func (e Event) Nack() error {
	if e.nackFn == nil {
		return fmt.Errorf("%w: nil nack function", value.ErrInvalidOperation)
	}

	return e.nackFn()
}
