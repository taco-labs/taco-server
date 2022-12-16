package entity

import (
	"time"

	"github.com/taco-labs/taco/go/common/analytics"
	"github.com/taco-labs/taco/go/utils"
	"github.com/uptrace/bun"
)

type Analytics struct {
	bun.BaseModel `bun:"table:analytics"`

	Id         string                   `json:"id"`
	EventType  analytics.EventType      `json:"event_type"`
	Payload    analytics.AnalyticsEvent `json:"payload,type:jsonb"`
	CreateTime time.Time                `json:"create_time"`
}

func NewAnalytics(createTime time.Time, payload analytics.AnalyticsEvent) Analytics {
	return Analytics{
		Id:         utils.MustNewUUID(),
		EventType:  payload.EventType(),
		Payload:    payload,
		CreateTime: createTime,
	}
}
