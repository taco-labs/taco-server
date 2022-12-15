package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/app/outbox"
	"github.com/taco-labs/taco/go/common/analytics"
	"github.com/taco-labs/taco/go/config"
	"github.com/taco-labs/taco/go/repository"
	"github.com/taco-labs/taco/go/service"
	"github.com/taco-labs/taco/go/utils"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"go.uber.org/zap"
	"gocloud.dev/pubsub/awssnssqs"
)

func RunOutbox(ctx context.Context, outboxConfig config.OutboxConfig, logger *zap.Logger, quit <-chan (os.Signal)) error {
	ctx = utils.SetLogger(ctx, logger)

	// Initialize analytics logger
	analytics.InitLogger(outboxConfig.Env)

	// Initialize aws sdk v2 session
	awsconf, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("Failed to load default aws config: %w", err)
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&search_path=%s",
		outboxConfig.Database.UserName,
		outboxConfig.Database.Password,
		outboxConfig.Database.Host,
		outboxConfig.Database.Port,
		outboxConfig.Database.Database,
		outboxConfig.Database.Schema,
	)

	// TODO (taekyeom) connection pool parameter?
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

	db := bun.NewDB(sqldb, pgdialect.New())

	if outboxConfig.Log.Query {
		hook := bundebug.NewQueryHook(bundebug.WithVerbose(false))
		db.AddQueryHook(hook)
	}

	transactor := app.NewDefaultTranscator(db)

	eventRepository := repository.NewEventRepository()

	sqsClient := sqs.NewFromConfig(awsconf)
	eventPublisher := awssnssqs.OpenSQSTopicV2(ctx, sqsClient, outboxConfig.EventTopic.Uri, &awssnssqs.TopicOptions{
		BodyBase64Encoding: awssnssqs.Never,
	})
	eventPublisherService := service.NewSqsPubService(eventPublisher)

	outboxApp, err := outbox.NewOutboxApp(
		outbox.WithTransactor(transactor),
		outbox.WithEventRepository(eventRepository),
		outbox.WithEventPublishService(eventPublisherService),
		outbox.WithTargetEventUriPrefix(outboxConfig.EventOutbox.EventUriPrefix),
		outbox.WithPollInterval(outboxConfig.EventOutbox.PollInterval),
		outbox.WithMaxMessages(outboxConfig.EventOutbox.MaxMessages),
	)
	if err != nil {
		return fmt.Errorf("failed to setup outbox app: %w", err)
	}

	if err := outboxApp.Start(ctx); err != nil {
		return fmt.Errorf("failed to start outbox app: %w", err)
	}

	<-quit
	fmt.Printf("shutting down outbox...")

	outboxApp.Shuwdown()

	eventPublisher.Shutdown(ctx)

	return nil
}
