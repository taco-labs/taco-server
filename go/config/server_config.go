package config

import (
	"context"

	"github.com/sethvargo/go-envconfig"
)

type ServerConfig struct {
	Log                LogConfig
	Database           DatabaseConfig
	SmsSender          SmsSenderConfig
	PaymentService     PaymentServiceConfig
	RouteService       RouteServiceConfig
	LocationService    LocationServiceConfig
	Backoffice         BackofficeConfig
	Firebase           FirebaseConfig
	NotificationTopic  TopicConfig          `env:",prefix=TACO_NOTIFICATION_"`
	TaxicallTopic      TopicConfig          `env:",prefix=TACO_TAXICALL_"`
	NotificationOutbox EventOutboxConfig    `env:",prefix=TACO_NOTIFICATION_OUTBOX_"`
	TaxicallOutbox     EventOutboxConfig    `env:",prefix=TACO_TAXICALL_OUTBOX_"`
	ImageUrlService    S3PresignedUrlConfig `env:",prefix=TACO_IMAGE_URL_SERVICE_"`

	TaxicallApp TaxicallAppConfig `env:",prefix=TACO_TAXICALL_APP_"`
	PushApp     PushAppConfig     `env:",prefix=TACO_PUSH_APP_"`
}

type TaxicallAppConfig struct {
	WorkerPoolConfig `env:",prefix=WORKER_"`
}

type PushAppConfig struct {
	WorkerPoolConfig `env:",prefix=WORKER_"`
}

func NewServerConfig(ctx context.Context) (ServerConfig, error) {
	config := ServerConfig{}

	err := envconfig.Process(ctx, &config)

	return config, err
}
