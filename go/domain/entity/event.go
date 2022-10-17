package entity

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/utils"
	"github.com/uptrace/bun"
)

const (
	MetadataKey_RetryCount = "retry_count"
	MetaDataKey_EventUri   = "event_uri"
	MetaDataKey_MessageId  = "message_id"
)

type Event struct {
	bun.BaseModel `bun:"table:event"`

	MessageId    string          `bun:"message_id,pk"`
	EventUri     string          `bun:"event_uri"`
	DelaySeconds int64           `bun:"delay_seconds"`
	Payload      json.RawMessage `bun:"payload,type:jsonb"`
	CreateTime   time.Time       `bun:"create_time"`
	RetryCount   int             `bun:"-"`
	ackFn        func() error
}

func (e Event) NewEventWithRetry() Event {
	return Event{
		MessageId:    utils.MustNewUUID(),
		EventUri:     e.EventUri,
		DelaySeconds: e.DelaySeconds,
		Payload:      e.Payload,
		CreateTime:   time.Now().UTC(),
		RetryCount:   e.RetryCount + 1,
	}
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

func (e Event) Ack() error {
	if e.ackFn == nil {
		return fmt.Errorf("%w: nil ack function", value.ErrInvalidOperation)
	}

	return e.ackFn()
}
