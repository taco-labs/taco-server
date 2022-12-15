package config

import (
	"context"

	"github.com/sethvargo/go-envconfig"
)

type OutboxConfig struct {
	Env         string `env:"TACO_ENV,default=dev"`
	Log         LogConfig
	Database    DatabaseConfig
	EventTopic  TopicConfig       `env:",prefix=TACO_EVENT_"`
	EventOutbox EventOutboxConfig `env:",prefix=TACO_EVENT_OUTBOX_"`
}

func NewOutboxConfig(ctx context.Context) (OutboxConfig, error) {
	config := OutboxConfig{}

	err := envconfig.Process(ctx, &config)

	return config, err
}
