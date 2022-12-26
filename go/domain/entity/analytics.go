package entity

import (
	"time"

	"github.com/taco-labs/taco/go/domain/value/analytics"
	"github.com/taco-labs/taco/go/utils"
	"github.com/uptrace/bun"
)

type Analytics struct {
	bun.BaseModel `bun:"table:analytics"`

	Id         string                   `bun:"id"`
	EventType  analytics.EventType      `bun:"event_type"`
	Payload    analytics.AnalyticsEvent `bun:"payload,type:jsonb"`
	CreateTime time.Time                `bun:"create_time"`
}

func NewAnalytics(createTime time.Time, payload analytics.AnalyticsEvent) Analytics {
	return Analytics{
		Id:         utils.MustNewUUID(),
		EventType:  payload.EventType(),
		Payload:    payload,
		CreateTime: createTime,
	}
}
