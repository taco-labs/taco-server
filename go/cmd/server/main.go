package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	firebase "firebase.google.com/go"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/dgraph-io/ristretto"
	"github.com/eko/gocache/v3/cache"
	"github.com/eko/gocache/v3/store"
	"github.com/panjf2000/ants/v2"
	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/app/driver"
	"github.com/taco-labs/taco/go/app/driversession"
	"github.com/taco-labs/taco/go/app/driversettlement"
	"github.com/taco-labs/taco/go/app/outbox"
	"github.com/taco-labs/taco/go/app/payment"
	"github.com/taco-labs/taco/go/app/push"
	"github.com/taco-labs/taco/go/app/taxicall"
	"github.com/taco-labs/taco/go/app/user"
	"github.com/taco-labs/taco/go/app/usersession"
	"github.com/taco-labs/taco/go/config"
	"github.com/taco-labs/taco/go/repository"
	"github.com/taco-labs/taco/go/server"
	backofficeserver "github.com/taco-labs/taco/go/server/backoffice"
	driverserver "github.com/taco-labs/taco/go/server/driver"
	userserver "github.com/taco-labs/taco/go/server/user"
	paypleextension "github.com/taco-labs/taco/go/server/user/extensions/payple"
	"github.com/taco-labs/taco/go/service"
	"github.com/taco-labs/taco/go/utils"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"go.uber.org/zap"
	"gocloud.dev/pubsub/awssnssqs"
	firebasepubsub "gocloud.dev/pubsub/firebase"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	config, err := config.NewServerConfig(ctx)
	if err != nil {
		fmt.Printf("Failed to initialize taco config: %+v\n", err)
		os.Exit(1)
	}

	var logger *zap.Logger
	if config.Env == "prod" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		fmt.Printf("Failed to initializae logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()
	ctx = utils.SetLogger(ctx, logger)

	// Initialize aws sdk v2 session
	awsconf, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Error("Failed to load default aws config", zap.Error(err))
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

	// TODO (taekyeom) connection pool parameter?
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

	db := bun.NewDB(sqldb, pgdialect.New())

	if config.Log.Query {
		hook := bundebug.NewQueryHook(bundebug.WithVerbose(false))
		db.AddQueryHook(hook)
	}

	transactor := app.NewDefaultTranscator(db)

	// Init repositories

	smsVerificationRepository := repository.NewSmsVerificationRepository()

	userRepository := repository.NewUserRepository()
	userSessionRepository := repository.NewUserSessionRepository()
	userPaymentRepository := repository.NewUserPaymentRepository()

	taxiCallRequestRepository := repository.NewTaxiCallRepository()

	driverRepository := repository.NewDriverRepository()
	driverLocationRepository := repository.NewDriverLocationRepository()
	driverSettlementAccountRepository := repository.NewDriverSettlementAccountRepository()
	driverSettlementRepository := repository.NewDriverSettlementRepository()
	driverSessionRepository := repository.NewDriverSessionRepository()

	eventRepository := repository.NewEventRepository()

	pushTokenRepository := repository.NewPushTokenRepository()

	// Init services

	smsSenderService := service.NewCoolSmsSenderService(
		config.SmsSender.Endpoint,
		config.SmsSender.SenderPhone,
		config.SmsSender.ApiKey,
		config.SmsSender.ApiSecret,
	)

	mapRouteService := service.NewNaverMapsRouteService(
		config.RouteService.Endpoint,
		config.RouteService.ApiKey,
		config.RouteService.ApiSecret,
	)

	locationService := service.NewKakaoLocationService(
		config.LocationService.Endpoint,
		config.LocationService.ApiSecret,
	)

	// TODO(taekyeom) Replace mock to real one
	payplePaymentService := service.NewPayplePaymentService(
		config.PaymentService.Endpoint,
		config.PaymentService.RefererDomain,
		config.PaymentService.ApiKey,
		config.PaymentService.ApiSecret,
	)

	var settlementAccountService service.SettlementAccountService
	switch config.SettlementAccountService.Type {
	case "mock":
		settlementAccountService = service.NewMockSettlementAccountService()
	case "payple":
		settlementAccountService = service.NewPaypleSettlemtnAccountService(
			config.SettlementAccountService.Endpoint,
			config.SettlementAccountService.ApiKey,
			config.SettlementAccountService.ApiSecret,
		)
	default:
		logger.Error(fmt.Sprintf("invalid settlement account service type '%s'", config.SettlementAccountService.Type))
		os.Exit(1)
	}

	eventStreamAntWorkerPool, err := ants.NewPool(config.EventStream.EventStreamWorkerPool.PoolSize,
		ants.WithPreAlloc(config.EventStream.EventStreamWorkerPool.PreAlloc),
	)
	if err != nil {
		logger.Error("failed to instantiate event stream ant worker pool", zap.Error(err))
		os.Exit(1)
	}
	eventStreamWorkerPool := service.NewAntWorkerPoolService(eventStreamAntWorkerPool)

	firebaseApp, err := firebase.NewApp(ctx, nil)
	if err != nil {
		logger.Error("failed to instantiate firebase", zap.Error(err))
		os.Exit(1)
	}
	messagingClient, err := firebaseApp.Messaging(ctx)
	if err != nil {
		logger.Error("failed to instantiate firebase cloud messaging client", zap.Error(err))
		os.Exit(1)
	}
	firebasepub := firebasepubsub.OpenFCMTopic(ctx, messagingClient, &firebasepubsub.TopicOptions{
		DryRun: config.Firebase.DryRun,
	})
	defer firebasepub.Shutdown(ctx)
	notificationService := service.NewFirebaseNotificationService(firebasepub)

	sqsClient := sqs.NewFromConfig(awsconf)
	eventSubscriber := awssnssqs.OpenSubscriptionV2(ctx, sqsClient, config.EventStream.EventTopic.Uri, &awssnssqs.SubscriptionOptions{
		WaitTime: time.Second,
		Raw:      true,
	})
	defer eventSubscriber.Shutdown(ctx)
	eventSubscriberService := service.NewSqsSubService(eventSubscriber)
	eventSubsriberStreamService := service.NewEventSubscriptionStreamService(eventSubscriberService, eventStreamWorkerPool)

	eventPublisher := awssnssqs.OpenSQSTopicV2(ctx, sqsClient, config.EventStream.EventTopic.Uri, &awssnssqs.TopicOptions{
		BodyBase64Encoding: awssnssqs.Never,
	})
	eventPublisherService := service.NewSqsPubService(eventPublisher)

	s3Client := s3.NewFromConfig(awsconf)
	presignedClient := s3.NewPresignClient(s3Client)
	s3ImagePresignedUrlService := service.NewS3ImagePresignedUrlService(
		presignedClient,
		config.ImageUrlService.Timeout,
		config.ImageUrlService.Bucket,
		config.ImageUrlService.BasePath,
	)

	// TODO(taekyeom) unify cache interface regardless of its method
	downloadImageUrlRistrettoCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: int64(10 * config.ImageUrlService.MaxCacheSizeEntires),
		MaxCost:     int64(config.ImageUrlService.MaxCacheSizeBytes),
		BufferItems: 64,
	})
	if err != nil {
		logger.Error("failed to setup download url ristretto cache", zap.Error(err))
		os.Exit(1)
	}
	downloadImageUrlCache := store.NewRistretto(downloadImageUrlRistrettoCache,
		store.WithExpiration(config.ImageUrlService.Timeout),
	)
	downloadImageUrlCacheManager := cache.New[string](downloadImageUrlCache)

	uploadImageUrlRistrettoCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: int64(10 * config.ImageUrlService.MaxCacheSizeEntires),
		MaxCost:     int64(config.ImageUrlService.MaxCacheSizeBytes),
		BufferItems: 64,
	})
	if err != nil {
		logger.Error("failed to setup upload url ristretto cache", zap.Error(err))
		os.Exit(1)
	}
	uploadImageUrlCache := store.NewRistretto(uploadImageUrlRistrettoCache,
		store.WithExpiration(config.ImageUrlService.Timeout),
	)
	uploadImageUrlCacheManager := cache.New[string](uploadImageUrlCache)

	cachedS3ImagePresignedUrlService := service.NewCachedUrlService(
		downloadImageUrlCacheManager,
		uploadImageUrlCacheManager,
		s3ImagePresignedUrlService)

	// Init apps
	userGetterDelegator := push.NewUserGetterDelegator()
	driverGetterDelegator := push.NewDriverGetterDelegator()

	pushApp, err := push.NewPushApp(
		push.WithTransactor(transactor),
		push.WithRouteService(mapRouteService),
		push.WithNotificationService(notificationService),
		push.WithPushTokenRepository(pushTokenRepository),
		push.WithUserGetterService(userGetterDelegator),
		push.WithDriverGetterService(driverGetterDelegator),
	)
	if err != nil {
		logger.Error("failed to setup push app", zap.Error(err))
		os.Exit(1)
	}

	taxicallApp, err := taxicall.NewTaxicallApp(
		taxicall.WithTransactor(transactor),
		taxicall.WithDriverLocationRepository(driverLocationRepository),
		taxicall.WithTaxiCallRequestRepository(taxiCallRequestRepository),
		taxicall.WithEventRepository(eventRepository),
		taxicall.WithRouteServie(mapRouteService),
		taxicall.WithLocationService(locationService),
		taxicall.WithUserGetterService(userGetterDelegator),
		taxicall.WithDriverGetterService(driverGetterDelegator),
	)
	if err != nil {
		logger.Error("failed to setup taxi call app", zap.Error(err))
		os.Exit(1)
	}

	paymentApp, err := payment.NewPaymentApp(
		payment.WithTransactor(transactor),
		payment.WithPaymentRepository(userPaymentRepository),
		payment.WithEventRepository(eventRepository),
		payment.WithPaymentService(payplePaymentService),
	)
	if err != nil {
		logger.Error("failed to setup user payment app", zap.Error(err))
		os.Exit(1)
	}

	outboxApp, err := outbox.NewOutboxApp(
		outbox.WithTransactor(transactor),
		outbox.WithEventRepository(eventRepository),
		outbox.WithEventPublishService(eventPublisherService),
		outbox.WithTargetEventUriPrefix(config.EventStream.EventOutbox.EventUriPrefix),
		outbox.WithPollInterval(config.EventStream.EventOutbox.PollInterval),
		outbox.WithMaxMessages(config.EventStream.EventOutbox.MaxMessages),
	)
	if err != nil {
		logger.Error("failed to setup outbox app", zap.Error(err))
		os.Exit(1)
	}

	userSessionApp, err := usersession.NewUserSessionApp(
		usersession.WithTransactor(transactor),
		usersession.WithUserSessionRepository(userSessionRepository),
	)
	if err != nil {
		logger.Error("failed to setup user session app", zap.Error(err))
		os.Exit(1)
	}

	driverSessionApp, err := driversession.NewDriverSessionApp(
		driversession.WithTransactor(transactor),
		driversession.WithDriverSessionRepository(driverSessionRepository),
	)
	if err != nil {
		logger.Error("failed to setup driver session app", zap.Error(err))
		os.Exit(1)
	}

	driverSettlementApp, err := driversettlement.NewDriverSettlementApp(
		driversettlement.WithTransactor(transactor),
		driversettlement.WithSettlementRepository(driverSettlementRepository),
	)
	if err != nil {
		logger.Error("failed to setup driver settlement app", zap.Error(err))
		os.Exit(1)
	}

	userApp, err := user.NewUserApp(
		user.WithTransactor(transactor),
		user.WithUserRepository(userRepository),
		user.WithSessionService(userSessionApp),
		user.WithSmsVerificationRepository(smsVerificationRepository),
		user.WithSmsSenderService(smsSenderService),
		user.WithMapRouteService(mapRouteService),
		user.WithLocationService(locationService),
		user.WithPushService(pushApp),
		user.WithTaxiCallService(taxicallApp),
		user.WithUserPaymentService(paymentApp),
	)
	if err != nil {
		logger.Error("failed to setup user app", zap.Error(err))
		os.Exit(1)
	}

	driverApp, err := driver.NewDriverApp(
		driver.WithTransactor(transactor),
		driver.WithDriverRepository(driverRepository),
		driver.WithSettlementAccountRepository(driverSettlementAccountRepository),
		driver.WithSessionService(driverSessionApp),
		driver.WithSmsSenderService(smsSenderService),
		driver.WithSmsVerificationRepository(smsVerificationRepository),
		driver.WithEventRepository(eventRepository),
		driver.WithPushService(pushApp),
		driver.WithTaxiCallService(taxicallApp),
		driver.WithImageUrlService(cachedS3ImagePresignedUrlService),
		driver.WithSettlementAccountService(settlementAccountService),
		driver.WithDriverSettlementService(driverSettlementApp),
	)
	if err != nil {
		logger.Error("failed to setup driver app", zap.Error(err))
		os.Exit(1)
	}

	// Run apps
	userGetterDelegator.Set(userApp)
	driverGetterDelegator.Set(driverApp)

	// Run subscription stream
	eventSubsriberStreamService.Add(pushApp)
	eventSubsriberStreamService.Add(taxicallApp)
	eventSubsriberStreamService.Add(paymentApp)
	eventSubsriberStreamService.Add(driverSettlementApp)

	eventSubsriberStreamService.Run(ctx)
	defer eventSubsriberStreamService.Shutdown(ctx)

	if err := outboxApp.Start(ctx); err != nil {
		logger.Error("failed to start outbox app", zap.Error(err))
		os.Exit(1)
	}
	defer outboxApp.Shuwdown()

	// Init middlewares
	loggerMiddleware := server.NewLoggerMiddleware(logger)

	userSessionMiddleware := userserver.NewSessionMiddleware(userSessionApp)

	driverSessionMiddleware := driverserver.NewSessionMiddleware(driverSessionApp)

	backofficeSessionMiddleware := backofficeserver.NewSessionMiddleware(config.Backoffice.Secret)

	// Init server extensions
	userServerPaypleExtension, err := paypleextension.NewPaypleExtension(
		paypleextension.WithPayplePaymentApp(userApp),
		paypleextension.WithDomain(config.PaymentService.RefererDomain),
		paypleextension.WithEnv(config.Env),
	)
	if err != nil {
		logger.Error("failed to setup payple extension", zap.Error(err))
		os.Exit(1)
	}

	// Init servers
	userServer, err := userserver.NewUserServer(
		userserver.WithEndpoint("0.0.0.0"),
		userserver.WithPort(18881),
		userserver.WithUserApp(userApp),
		userserver.WithMiddleware(loggerMiddleware.Process),
		userserver.WithMiddleware(userSessionMiddleware.Get()),
		userserver.WithMiddleware(userserver.UserIdChecker),
		userserver.WithExtension(userServerPaypleExtension.Apply),
	)
	if err != nil {
		logger.Error("failed to setup user server", zap.Error(err))
		os.Exit(1)
	}
	defer userServer.Stop(ctx)

	driverServer, err := driverserver.NewDriverServer(
		driverserver.WithEndpoint("0.0.0.0"),
		driverserver.WithPort(18882),
		driverserver.WithDriverApp(driverApp),
		driverserver.WithMiddleware(loggerMiddleware.Process),
		driverserver.WithMiddleware(driverSessionMiddleware.Get()),
		driverserver.WithMiddleware(driverserver.DriverIdChecker),
	)
	if err != nil {
		logger.Error("failed to setup driver server", zap.Error(err))
		os.Exit(1)
	}
	defer driverServer.Stop(ctx)

	backofficeServer, err := backofficeserver.NewBackofficeServer(
		backofficeserver.WithEndpoint("0.0.0.0"),
		backofficeserver.WithPort(18883),
		backofficeserver.WithDriverApp(driverApp),
		backofficeserver.WithUserApp(userApp),
		backofficeserver.WithMiddleware(loggerMiddleware.Process),
		backofficeserver.WithMiddleware(backofficeSessionMiddleware.Get()),
	)
	if err != nil {
		logger.Error("failed to setup backoffice server", zap.Error(err))
		os.Exit(1)
	}
	defer backofficeServer.Stop(ctx)

	go func() {
		if err := userServer.Run(ctx); err != nil && err != http.ErrServerClosed {
			logger.Fatal("shutting down user server", zap.Error(err))
		}
	}()

	go func() {
		if err := driverServer.Run(ctx); err != nil && err != http.ErrServerClosed {
			logger.Fatal("shutting down driver server", zap.Error(err))
		}
	}()

	go func() {
		if err := backofficeServer.Run(ctx); err != nil && err != http.ErrServerClosed {
			logger.Fatal("shutting down backoffice server", zap.Error(err))
		}
	}()

	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	fmt.Println("shutting down [Taco-Backend] service... because of interrupt")
	cancel()
}
