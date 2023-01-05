package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	firebase "firebase.google.com/go"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
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
	"github.com/taco-labs/taco/go/app/payment"
	"github.com/taco-labs/taco/go/app/push"
	"github.com/taco-labs/taco/go/app/taxicall"
	"github.com/taco-labs/taco/go/app/user"
	"github.com/taco-labs/taco/go/app/usersession"
	"github.com/taco-labs/taco/go/config"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/repository"
	"github.com/taco-labs/taco/go/server"
	backofficeserver "github.com/taco-labs/taco/go/server/backoffice"
	driverserver "github.com/taco-labs/taco/go/server/driver"
	driverpaypleextension "github.com/taco-labs/taco/go/server/driver/extensions/payple"
	userserver "github.com/taco-labs/taco/go/server/user"
	userpaypleextension "github.com/taco-labs/taco/go/server/user/extensions/payple"
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

func RunServer(ctx context.Context, serverConfig config.ServerConfig, logger *zap.Logger, quit <-chan (os.Signal)) error {
	ctx = utils.SetLogger(ctx, logger)

	// Initialize aws sdk v2 session
	awsconf, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load default aws config: %w", err)
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&search_path=%s",
		serverConfig.Database.UserName,
		serverConfig.Database.Password,
		serverConfig.Database.Host,
		serverConfig.Database.Port,
		serverConfig.Database.Database,
		serverConfig.Database.Schema,
	)

	// TODO (taekyeom) connection pool parameter?
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

	db := bun.NewDB(sqldb, pgdialect.New())

	if serverConfig.Log.Query {
		hook := bundebug.NewQueryHook(bundebug.WithVerbose(false))
		db.AddQueryHook(hook)
	}

	transactor := app.NewDefaultTranscator(db)

	// Init repositories

	smsVerificationRepository := repository.NewSmsVerificationRepository()

	userRepository := repository.NewUserRepository()
	userSessionRepository := repository.NewUserSessionRepository()
	userPaymentRepository := repository.NewUserPaymentRepository()
	referralRepository := repository.NewReferralRepository()

	taxiCallRequestRepository := repository.NewTaxiCallRepository()

	driverRepository := repository.NewDriverRepository()
	driverLocationRepository := repository.NewDriverLocationRepository()
	driverSettlementAccountRepository := repository.NewDriverSettlementAccountRepository()
	driverSettlementRepository := repository.NewDriverSettlementRepository()
	driverSessionRepository := repository.NewDriverSessionRepository()

	eventRepository := repository.NewEventRepository()
	analyticsRepository := repository.NewAnalyticsRepository()

	pushTokenRepository := repository.NewPushTokenRepository()

	// Init services
	var smsVerificationService service.SmsVerificationSenderService
	switch serverConfig.SmsSender.Type {
	case "mock":
		smsVerificationService = service.NewMockSmsSenderService()
	case "coolsms":
		smsVerificationService = service.NewCoolSmsSenderService(
			serverConfig.SmsSender.Endpoint,
			serverConfig.SmsSender.SenderPhone,
			serverConfig.SmsSender.ApiKey,
			serverConfig.SmsSender.ApiSecret,
		)
	}

	var mapService service.MapService
	switch serverConfig.MapService.Type {
	case "mock":
		mapService = service.NewMockMapService()
	case "nhn":
		mapService = service.NewNhnMapsService(
			serverConfig.MapService.Endpoint,
			serverConfig.MapService.ApiKey,
		)
	}

	// TODO(taekyeom) Replace mock to real one
	payplePaymentService := service.NewPayplePaymentService(
		serverConfig.PaymentService.Endpoint,
		serverConfig.PaymentService.RefererDomain,
		serverConfig.PaymentService.ApiKey,
		serverConfig.PaymentService.ApiSecret,
	)

	var settlementAccountService service.SettlementAccountService
	switch serverConfig.SettlementAccountService.Type {
	case "mock":
		settlementAccountService = service.NewMockSettlementAccountService(
			serverConfig.SettlementAccountService.Endpoint,
			serverConfig.SettlementAccountService.ApiKey,
			serverConfig.SettlementAccountService.ApiSecret,
			serverConfig.SettlementAccountService.WebhookUrl,
		)
	case "payple":
		settlementAccountService = service.NewPaypleSettlemtnAccountService(
			serverConfig.SettlementAccountService.Endpoint,
			serverConfig.SettlementAccountService.ApiKey,
			serverConfig.SettlementAccountService.ApiSecret,
		)
	default:
		return fmt.Errorf(fmt.Sprintf("invalid settlement account service type '%s'", serverConfig.SettlementAccountService.Type))
	}

	eventStreamAntWorkerPool, err := ants.NewPool(serverConfig.EventStream.EventStreamWorkerPool.PoolSize,
		ants.WithPreAlloc(serverConfig.EventStream.EventStreamWorkerPool.PreAlloc),
	)
	if err != nil {
		return fmt.Errorf("failed to instantiate event stream ant worker pool: %w", err)
	}
	eventStreamWorkerPool := service.NewAntWorkerPoolService(eventStreamAntWorkerPool)

	firebaseApp, err := firebase.NewApp(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to instantiate firebase: %w", err)
	}

	messagingClient, err := firebaseApp.Messaging(ctx)
	if err != nil {
		return fmt.Errorf("failed to instantiate firebase cloud messaging client: %w", err)
	}
	firebasepub := firebasepubsub.OpenFCMTopic(ctx, messagingClient, &firebasepubsub.TopicOptions{
		DryRun: serverConfig.Firebase.DryRun,
	})
	notificationService := service.NewFirebaseNotificationService(firebasepub)

	sqsClient := sqs.NewFromConfig(awsconf)
	eventSubscriber := awssnssqs.OpenSubscriptionV2(ctx, sqsClient, serverConfig.EventStream.EventTopic.Uri, &awssnssqs.SubscriptionOptions{
		WaitTime: time.Second,
		Raw:      true,
	})
	eventSubscriberService := service.NewSqsSubService(eventSubscriber)
	eventSubsriberStreamService := service.NewEventSubscriptionStreamService(eventSubscriberService, eventStreamWorkerPool)

	s3Client := s3.NewFromConfig(awsconf)
	presignedClient := s3.NewPresignClient(s3Client)
	s3ImagePresignedUrlService := service.NewS3ImagePresignedUrlService(
		presignedClient,
		serverConfig.ImageUploadUrlService.Timeout,
		serverConfig.ImageUploadUrlService.Bucket,
		serverConfig.ImageUploadUrlService.BasePath,
	)

	uploadImageUrlRistrettoCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: int64(10 * serverConfig.ImageUploadUrlService.MaxCacheSizeEntires),
		MaxCost:     int64(serverConfig.ImageUploadUrlService.MaxCacheSizeBytes),
		BufferItems: 64,
	})
	if err != nil {
		return fmt.Errorf("failed to setup upload url ristretto cache: %w", err)
	}
	uploadImageUrlCache := store.NewRistretto(uploadImageUrlRistrettoCache,
		store.WithExpiration(serverConfig.ImageUploadUrlService.Timeout),
	)
	uploadImageUrlCacheManager := cache.New[string](uploadImageUrlCache)

	cachedS3ImagePresignedUploadUrlService := service.NewCachedUrlService(
		uploadImageUrlCacheManager,
		s3ImagePresignedUrlService)

	publicS3ImageDownloadUrlService := service.NewS3PublicAccessUrlService(
		serverConfig.ImageDownloadUrlService.RegionalDomain,
		serverConfig.ImageDownloadUrlService.BasePath)

	kmsClient := kms.NewFromConfig(awsconf)
	kmsEncryptionService := service.NewAwsKMSEncryptionService(kmsClient, serverConfig.EncryptionService.KeyId)

	userServiceRegionChecker, err := service.NewStaticServiceRegionChecker(serverConfig.ServiceRegion.UserServiceRegions, serverConfig.ServiceRegion.UserServiceRegions)
	if err != nil {
		return fmt.Errorf("failed to setup user service region checker: %w", err)
	}

	driverServiceRegionChecker, err := service.NewStaticServiceRegionChecker(value.SupportedServiceRegionList, serverConfig.ServiceRegion.DriverServiceRegions)
	if err != nil {
		return fmt.Errorf("failed to setup driver service region checker: %w", err)
	}

	// Init apps
	userAppDelegator := user.NewUserAppDelegator()
	driverAppDelegator := driver.NewDriverAppDelegator()

	pushApp, err := push.NewPushApp(
		push.WithTransactor(transactor),
		push.WithNotificationService(notificationService),
		push.WithPushTokenRepository(pushTokenRepository),
		push.WithUserGetterService(userAppDelegator),
		push.WithDriverGetterService(driverAppDelegator),
	)
	if err != nil {
		return fmt.Errorf("failed to setup push app: %w", err)
	}

	driverSettlementApp, err := driversettlement.NewDriverSettlementApp(
		driversettlement.WithTransactor(transactor),
		driversettlement.WithSettlementRepository(driverSettlementRepository),
		driversettlement.WithEventRepository(eventRepository),
		driversettlement.WithSettlementAccountService(settlementAccountService),
		driversettlement.WithAnalyticsRepository(analyticsRepository),
	)
	if err != nil {
		return fmt.Errorf("failed to setup driver settlement app: %w", err)
	}

	paymentApp, err := payment.NewPaymentApp(
		payment.WithTransactor(transactor),
		payment.WithPaymentRepository(userPaymentRepository),
		payment.WithEventRepository(eventRepository),
		payment.WithPaymentService(payplePaymentService),
		payment.WithReferralRepository(referralRepository),
		payment.WithAnalyticsRepository(analyticsRepository),
		payment.WithDriverSettlementService(driverSettlementApp),
	)
	if err != nil {
		return fmt.Errorf("failed to setup user payment app: %w", err)
	}

	taxicallApp, err := taxicall.NewTaxicallApp(
		taxicall.WithTransactor(transactor),
		taxicall.WithDriverLocationRepository(driverLocationRepository),
		taxicall.WithTaxiCallRequestRepository(taxiCallRequestRepository),
		taxicall.WithEventRepository(eventRepository),
		taxicall.WithMapService(mapService),
		taxicall.WithUserGetterService(userAppDelegator),
		taxicall.WithDriverGetterService(driverAppDelegator),
		taxicall.WithPaymentAppService(paymentApp),
		taxicall.WithAnalyticsRepository(analyticsRepository),
		taxicall.WithUserServiceRegionChecker(userServiceRegionChecker),
	)
	if err != nil {
		return fmt.Errorf("failed to setup taxi call app: %w", err)
	}

	userSessionApp, err := usersession.NewUserSessionApp(
		usersession.WithTransactor(transactor),
		usersession.WithUserSessionRepository(userSessionRepository),
	)
	if err != nil {
		return fmt.Errorf("failed to setup user session app: %w", err)
	}

	driverSessionApp, err := driversession.NewDriverSessionApp(
		driversession.WithTransactor(transactor),
		driversession.WithDriverSessionRepository(driverSessionRepository),
	)
	if err != nil {
		return fmt.Errorf("failed to setup driver session app: %w", err)
	}

	userApp, err := user.NewUserApp(
		user.WithTransactor(transactor),
		user.WithUserRepository(userRepository),
		user.WithSessionService(userSessionApp),
		user.WithSmsVerificationRepository(smsVerificationRepository),
		user.WithSmsSenderService(smsVerificationService),
		user.WithMapService(mapService),
		user.WithPushService(pushApp),
		user.WithTaxiCallService(taxicallApp),
		user.WithUserPaymentService(paymentApp),
		user.WithAnalyticsRepository(analyticsRepository),
		user.WithServiceRegionChecker(userServiceRegionChecker),
	)
	if err != nil {
		return fmt.Errorf("failed to setup user app: %w", err)
	}

	driverApp, err := driver.NewDriverApp(
		driver.WithTransactor(transactor),
		driver.WithDriverRepository(driverRepository),
		driver.WithSettlementAccountRepository(driverSettlementAccountRepository),
		driver.WithSessionService(driverSessionApp),
		driver.WithSmsSenderService(smsVerificationService),
		driver.WithSmsVerificationRepository(smsVerificationRepository),
		driver.WithEventRepository(eventRepository),
		driver.WithPushService(pushApp),
		driver.WithTaxiCallService(taxicallApp),
		driver.WithImageUploadUrlService(cachedS3ImagePresignedUploadUrlService),
		driver.WithImageDownloadUrlService(publicS3ImageDownloadUrlService),
		driver.WithSettlementAccountService(settlementAccountService),
		driver.WithDriverSettlementService(driverSettlementApp),
		driver.WithUserPaymentAppService(paymentApp),
		driver.WithEncryptionService(kmsEncryptionService),
		driver.WithAnalyticsRepository(analyticsRepository),
		driver.WithServiceRegionChecker(driverServiceRegionChecker),
	)
	if err != nil {
		return fmt.Errorf("failed to setup driver app: %w", err)
	}

	// Run apps
	userAppDelegator.Set(userApp)
	driverAppDelegator.Set(driverApp)

	// Run subscription stream
	eventSubsriberStreamService.Add(pushApp)
	eventSubsriberStreamService.Add(taxicallApp)
	eventSubsriberStreamService.Add(paymentApp)
	eventSubsriberStreamService.Add(driverSettlementApp)

	eventSubsriberStreamService.Run(ctx)

	// Init middlewares
	loggerMiddleware := server.NewLoggerMiddleware(logger)

	userSessionMiddleware := userserver.NewSessionMiddleware(userSessionApp)

	driverSessionMiddleware := driverserver.NewSessionMiddleware(driverSessionApp)

	backofficeSessionMiddleware := backofficeserver.NewSessionMiddleware(serverConfig.Backoffice.Secret)

	// Init server extensions
	userServerPaypleExtension, err := userpaypleextension.NewPaypleExtension(
		userpaypleextension.WithPayplePaymentApp(userApp),
		userpaypleextension.WithDomain(serverConfig.PaymentService.RefererDomain),
		userpaypleextension.WithEnv(serverConfig.Env),
	)
	if err != nil {
		return fmt.Errorf("failed to setup user payple extension: %w", err)
	}

	driverServerPaypleExtension, err := driverpaypleextension.NewPaypleExtension(
		driverpaypleextension.WithDriverSEttlementApp(driverSettlementApp),
	)
	if err != nil {
		return fmt.Errorf("failed to setup driver payple extension: %w", err)
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
		return fmt.Errorf("failed to setup user server: %w", err)
	}

	driverServer, err := driverserver.NewDriverServer(
		driverserver.WithEndpoint("0.0.0.0"),
		driverserver.WithPort(18882),
		driverserver.WithDriverApp(driverApp),
		driverserver.WithMiddleware(loggerMiddleware.Process),
		driverserver.WithMiddleware(driverSessionMiddleware.Get()),
		driverserver.WithMiddleware(driverserver.DriverIdChecker),
		driverserver.WithExtension(driverServerPaypleExtension.Apply),
	)
	if err != nil {
		return fmt.Errorf("failed to setup driver server: %w", err)
	}

	backofficeServer, err := backofficeserver.NewBackofficeServer(
		backofficeserver.WithEndpoint("0.0.0.0"),
		backofficeserver.WithPort(18883),
		backofficeserver.WithDriverApp(driverApp),
		backofficeserver.WithUserApp(userApp),
		backofficeserver.WithTaxicallApp(taxicallApp),
		backofficeserver.WithMiddleware(loggerMiddleware.Process),
		backofficeserver.WithMiddleware(backofficeSessionMiddleware.Get()),
	)
	if err != nil {
		return fmt.Errorf("failed to setup backoffice server: %w", err)
	}

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

	<-quit
	fmt.Printf("shutting down taco backend...")

	eventSubscriber.Shutdown(ctx)
	eventSubsriberStreamService.Shutdown(ctx)

	firebasepub.Shutdown(ctx)

	userServer.Stop(ctx)
	driverServer.Stop(ctx)
	backofficeServer.Stop(ctx)

	logger.Sync()

	return nil
}
