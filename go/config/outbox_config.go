package config

import (
	"context"

	"github.com/sethvargo/go-envconfig"
)

type OutboxConfig struct {
	Log                    LogConfig
	Database               DatabaseConfig
	NotificationEventTopic EventOutboxConfig `env:",prefix=TACO_NOTIFICATION_OUTBOX_"`
	TaxicallEventTopic     EventOutboxConfig `env:",prefix=TACO_TAXICALL_OUTBOX_"`
}

func NewOutboxConfig(ctx context.Context) (OutboxConfig, error) {
	config := OutboxConfig{}

	err := envconfig.Process(ctx, &config)

	return config, err
}
