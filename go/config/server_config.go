package config

import (
	"context"

	"github.com/sethvargo/go-envconfig"
)

type ServerConfig struct {
	Env                      string `env:"TACO_ENV"`
	Log                      LogConfig
	Database                 DatabaseConfig
	SmsSender                SmsSenderConfig
	PaymentService           PaymentServiceConfig
	SettlementAccountService SettlementAccountServiceConfig
	RouteService             RouteServiceConfig
	LocationService          LocationServiceConfig
	Backoffice               BackofficeConfig
	Firebase                 FirebaseConfig
	EventStream              EventStreamConfig
	ImageUrlService          S3PresignedUrlConfig `env:",prefix=TACO_IMAGE_URL_SERVICE_"`
}

type EventStreamConfig struct {
	EventTopic            TopicConfig       `env:",prefix=TACO_EVENT_"`
	EventOutbox           EventOutboxConfig `env:",prefix=TACO_EVENT_OUTBOX_"`
	EventStreamWorkerPool WorkerPoolConfig  `env:",prefix=TACO_EVENT_STREAM_WORKER_"`
}

func NewServerConfig(ctx context.Context) (ServerConfig, error) {
	config := ServerConfig{}

	err := envconfig.Process(ctx, &config)

	return config, err
}
