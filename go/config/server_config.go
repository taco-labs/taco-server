package config

import (
	"context"

	"github.com/sethvargo/go-envconfig"
)

type ServerConfig struct {
	Log                    LogConfig
	Database               DatabaseConfig
	SmsSender              SmsSenderConfig
	PaymentService         PaymentServiceConfig
	RouteService           RouteServiceConfig
	LocationService        LocationServiceConfig
	Backoffice             BackofficeConfig
	Firebase               FirebaseConfig
	NotificationSubscriber EventTopicSubscriberConfig `env:",prefix=TACO_NOTIFICATION_SUBSCRIBER_"`
	NotificationPublisher  TopicConfig                `env:",prefix=TACO_NOTIFICATION_PUBLISHER_"`
	TaxicallSubscriber     EventTopicSubscriberConfig `env:",prefix=TACO_TAXICALL_SUBSCRIBER_"`
	TaxicallPublisher      TopicConfig                `env:",prefix=TACO_TAXICALL_PUBLISHER_"`
}

func NewServerConfig(ctx context.Context) (ServerConfig, error) {
	config := ServerConfig{}

	err := envconfig.Process(ctx, &config)

	return config, err
}
