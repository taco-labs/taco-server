package config

import (
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
	Endpoint      string `env:"TACO_PAYMENT_SERVICE_ENDPOINT,required"`
	ApiKey        string `env:"TACO_PAYMENT_SERVICE_API_KEY,required"`
	ApiSecret     string `env:"TACO_PAYMENT_SERVICE_API_SECRET,required"`
	RefererDomain string `env:"TACO_PAYMENT_SERVICE_REFERER_DOMAIN,required"`
}

type SettlementAccountServiceConfig struct {
	Type      string `env:"TACO_SETTLEMENT_ACCOUNT_SERVICE_TYPE,default=mock"`
	Endpoint  string `env:"TACO_SETTLEMENT_ACCOUNT_SERVICE_ENDPOINT"`
	ApiKey    string `env:"TACO_SETTLEMENT_ACCOUNT_SERVICE_API_KEY"`
	ApiSecret string `env:"TACO_SETTLEMENT_ACCOUNT_SERVICE_API_SECRET"`
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
	EventUriPrefix string        `env:"EVENT_URI_PREFIX,required"`
	PollInterval   time.Duration `env:"POLL_INTERVAL,required"`
	MaxMessages    int           `env:"MAX_MESSAGES,required"`
}

type TopicConfig struct {
	Uri string `env:"TOPIC_URI,required"`
}

type WorkerPoolConfig struct {
	PoolSize int  `env:"POOL_SIZE,default=1"`
	PreAlloc bool `env:"PRE_ALLOC,default=false"`
}

type S3PresignedUrlConfig struct {
	Timeout             time.Duration `env:"TIMEOUT,required"`
	Bucket              string        `env:"BUCKET,required"`
	BasePath            string        `env:"BASE_PATH,required"`
	MaxCacheSizeBytes   int           `env:"MAX_CACHE_SIZE_BYTES,required"`
	MaxCacheSizeEntires int           `env:"MAX_CACHE_SIZE_ENTRIES,required"`
}
