package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/awssnssqs"

	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/app/outbox"
	"github.com/taco-labs/taco/go/config"
	"github.com/taco-labs/taco/go/repository"
	"github.com/taco-labs/taco/go/service"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	config, err := config.NewOutboxConfig(ctx)
	if err != nil {
		fmt.Println("Failed to initialize taco config: ", err)
		os.Exit(1)
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&search_path=%s",
		config.Database.UserName,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.Database,
		config.Database.Schema,
	)

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

	db := bun.NewDB(sqldb, pgdialect.New())

	if config.Log.Query {
		hook := bundebug.NewQueryHook(bundebug.WithVerbose(true))
		db.AddQueryHook(hook)
	}

	transactor := app.NewDefaultTranscator(db)

	eventRepository := repository.NewEventRepository()

	notificationPublishTopic, err := pubsub.OpenTopic(ctx, config.NotificationEventTopic.Topic.GetSqsUri())
	if err != nil {
		fmt.Println("Failed to initialize notification sqs publish topic: ", err)
		os.Exit(1)
	}
	defer notificationPublishTopic.Shutdown(ctx)
	notificationPublishService := service.NewSqsPubService(notificationPublishTopic)

	notificationOutboxApp, err := outbox.NewOutboxApp(
		outbox.WithTransactor(transactor),
		outbox.WithEventRepository(eventRepository),
		outbox.WithEventPublishService(notificationPublishService),
		outbox.WithTargetEventUirs(config.NotificationEventTopic.EventUris),
		outbox.WithPollInterval(config.NotificationEventTopic.PollInterval),
		outbox.WithMaxMessages(config.NotificationEventTopic.MaxMessages),
	)
	if err != nil {
		fmt.Println("Failed to initialize notification outbox app: ", err)
		os.Exit(1)
	}

	if err := notificationOutboxApp.Start(ctx); err != nil {
		fmt.Println("Failed to start notification outbox app: ", err)
		os.Exit(1)
	}
	defer notificationOutboxApp.Shuwdown()

	taxicallPublishTopic, err := pubsub.OpenTopic(ctx, config.TaxicallEventTopic.Topic.GetSqsUri())
	if err != nil {
		fmt.Println("Failed to initialize taxicall sqs publish topic: ", err)
		os.Exit(1)
	}
	defer taxicallPublishTopic.Shutdown(ctx)
	taxicallPublishService := service.NewSqsPubService(taxicallPublishTopic)

	taxicallOutboxApp, err := outbox.NewOutboxApp(
		outbox.WithTransactor(transactor),
		outbox.WithEventRepository(eventRepository),
		outbox.WithEventPublishService(taxicallPublishService),
		outbox.WithTargetEventUirs(config.TaxicallEventTopic.EventUris),
		outbox.WithPollInterval(config.TaxicallEventTopic.PollInterval),
		outbox.WithMaxMessages(config.TaxicallEventTopic.MaxMessages),
	)
	if err != nil {
		fmt.Println("Failed to initialize taxicall outbox app: ", err)
		os.Exit(1)
	}

	if err := taxicallOutboxApp.Start(ctx); err != nil {
		fmt.Println("Failed to start taxicall outbox app: ", err)
		os.Exit(1)
	}
	defer taxicallOutboxApp.Shuwdown()

	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	fmt.Println("shutting down [Outbox] service... because of interrupt")
	cancel()
}
