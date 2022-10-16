package config

import (
	"context"

	"github.com/sethvargo/go-envconfig"
)

type ServerConfig struct {
	Log             LogConfig
	Database        DatabaseConfig
	SmsSender       SmsSenderConfig
	PaymentService  PaymentServiceConfig
	RouteService    RouteServiceConfig
	LocationService LocationServiceConfig
	Backoffice      BackofficeConfig
	Firebase        FirebaseConfig
}

func NewServerConfig(ctx context.Context) (ServerConfig, error) {
	config := ServerConfig{}

	err := envconfig.Process(ctx, &config)

	return config, err
}
