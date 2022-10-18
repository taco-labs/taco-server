package config

import (
	"fmt"
	"time"
)

type LogConfig struct {
	Query bool `env:"TACO_ENABLE_QUERY_DEBUG_LOG,default=true"`
}

type DatabaseConfig struct {
	Host     string `env:"TACO_DATABASE_HOST,required"`
	Port     int    `env:"TACO_DATABASE_PORT,required"`
	Database string `env:"TACO_DATABASE_DATABASE,required"`
	UserName string `env:"TACO_DATABASE_USERNAME,required"`
	Password string `env:"TACO_DATABASE_PASSWORD,required"`
	Schema   string `env:"TACO_DATABASE_SCHEMA,required"`
}

type SmsSenderConfig struct {
	Endpoint    string `env:"TACO_SMS_SENDER_ENDPOINT,required"`
	SenderPhone string `env:"TACO_SMS_SENDER_PHONE,required"`
	ApiKey      string `env:"TACO_SMS_SENDER_API_KEY,required"`
	ApiSecret   string `env:"TACO_SMS_SENDER_API_SECRET,required"`
}

type PaymentServiceConfig struct {
	Endpoint  string `env:"TACO_PAYMENT_SERVICE_ENDPOINT,required"`
	ApiSecret string `env:"TACO_PAYMENT_SERVICE_API_SECRET,required"`
}

type RouteServiceConfig struct {
	Endpoint  string `env:"TACO_ROUTE_SERVICE_ENDPOINT,required"`
	ApiKey    string `env:"TACO_ROUTE_SERVICE_API_KEY,required"`
	ApiSecret string `env:"TACO_ROUTE_SERVICE_API_SECRET,required"`
}

type LocationServiceConfig struct {
	Endpoint  string `env:"TACO_LOCATION_SERVICE_ENDPOINT,required"`
	ApiSecret string `env:"TACO_LOCATION_SERVICE_API_SECRET,required"`
}

type BackofficeConfig struct {
	Secret string `env:"TACO_BACKOFFICE_SECRET,required"`
}

type FirebaseConfig struct {
	DryRun bool `env:"TACO_FIREBASE_DRY_RUN,default=true"`
}

type EventOutboxConfig struct {
	Topic        TopicConfig   `env:",prefix="`
	EventUris    []string      `env:"EVENT_URIS,required"`
	PollInterval time.Duration `env:"POLL_INTERVAL,required"`
	MaxMessages  int           `env:"MAX_MESSAGES,required"`
}

type EventTopicSubscriberConfig struct {
	Topic        TopicConfig   `env:",prefix="`
	PollInterval time.Duration `env:"POLL_INTERVAL,required"`
}

type TopicConfig struct {
	Uri string `env:"TOPIC_URI,required"`
}

func (e TopicConfig) GetSqsUri() string {
	return fmt.Sprintf("awssqs://%s?awssdk=v2", e.Uri)
}
