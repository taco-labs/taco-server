package config

import (
	"context"

	"github.com/sethvargo/go-envconfig"
)

type OutboxConfig struct {
	Log                    LogConfig
	Database               DatabaseConfig
	NotificationEventTopic EventTopicConfig `env:",prefix=TACO_NOTIFICATION_OUTBOX_"`
}

func NewOutboxConfig(ctx context.Context) (OutboxConfig, error) {
	config := OutboxConfig{}

	err := envconfig.Process(ctx, &config)

	return config, err
}
