package config

import (
	"github.com/kelseyhightower/envconfig"
)

var Config TacoConfig

type TacoConfig struct {
	Log             LogConfig
	Database        DatabaseConfig
	SmsSender       SmsSenderConfig
	PaymentService  PaymentServiceConfig
	RouteService    RouteServiceConfig
	LocationService LocationServiceConfig
	Backoffice      BackofficeConfig
	Firebase        FirebaseConfig
}

type LogConfig struct {
	Query bool `envconfig:"TACO_ENABLE_QUERY_DEBUG_LOG" default:"false"`
}

type DatabaseConfig struct {
	Host     string `envconfig:"TACO_DATABASE_HOST" required:"true"`
	Port     int    `envconfig:"TACO_DATABASE_PORT" required:"true"`
	Database string `envconfig:"TACO_DATABASE_DATABASE" required:"true"`
	UserName string `envconfig:"TACO_DATABASE_USERNAME" required:"true"`
	Password string `envconfig:"TACO_DATABASE_PASSWORD" required:"true"`
	Schema   string `envconfig:"TACO_DATABASE_SCHEMA" required:"true"`
}

type SmsSenderConfig struct {
	Endpoint    string `envconfig:"TACO_SMS_SENDER_ENDPOINT" required:"true"`
	SenderPhone string `envconfig:"TACO_SMS_SENDER_ENDPOINT" required:"true"`
	ApiKey      string `envconfig:"TACO_SMS_SENDER_ENDPOINT" required:"true"`
	ApiSecret   string `envconfig:"TACO_SMS_SENDER_API_SECRET" required:"true"`
}

type PaymentServiceConfig struct {
	Endpoint  string `envconfig:"TACO_PAYMENT_SERVICE_ENDPOINT" required:"true"`
	ApiSecret string `envconfig:"TACO_PAYMENT_SERVICE_API_SECRET" required:"true"`
}

type RouteServiceConfig struct {
	Endpoint  string `envconfig:"TACO_ROUTE_SERVICE_ENDPOINT" required:"true"`
	ApiKey    string `envconfig:"TACO_ROUTE_SERVICE_API_KEY" required:"true"`
	ApiSecret string `envconfig:"TACO_ROUTE_SERVICE_API_SECRET" required:"true"`
}

type LocationServiceConfig struct {
	Endpoint  string `envconfig:"TACO_LOCATION_SERVICE_ENDPOINT" required:"true"`
	ApiSecret string `envconfig:"TACO_LOCATION_SERVICE_API_SECRET" required:"true"`
}

type BackofficeConfig struct {
	Secret string `envconfig:"TACO_BACKOFFICE_SECRET" required:"true"`
}

type FirebaseConfig struct {
}

func NewTacoConfig() (TacoConfig, error) {
	config := TacoConfig{}

	err := envconfig.Process("taco", &config)

	return config, err
}
