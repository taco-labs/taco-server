package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	firebase "firebase.google.com/go"
	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/app/driver"
	"github.com/taco-labs/taco/go/app/driversession"
	"github.com/taco-labs/taco/go/app/outbox"
	"github.com/taco-labs/taco/go/app/push"
	"github.com/taco-labs/taco/go/app/taxicall"
	"github.com/taco-labs/taco/go/app/user"
	"github.com/taco-labs/taco/go/app/usersession"
	"github.com/taco-labs/taco/go/config"
	"github.com/taco-labs/taco/go/repository"
	backofficeserver "github.com/taco-labs/taco/go/server/backoffice"
	driverserver "github.com/taco-labs/taco/go/server/driver"
	userserver "github.com/taco-labs/taco/go/server/user"
	"github.com/taco-labs/taco/go/service"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/awssnssqs"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	config, err := config.NewServerConfig(ctx)
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

	fileUploadService := service.NewMockFileUploadService()

	// TODO(taekyeom) Replace mock to real one
	tossPaymentService := service.NewTossPaymentService(
		config.PaymentService.Endpoint,
		config.PaymentService.ApiSecret,
	)

	firebaseApp, err := firebase.NewApp(ctx, nil)
	if err != nil {
		fmt.Println("Failed to instantiate firebase: ", err)
		os.Exit(1)
	}

	messagingClient, err := firebaseApp.Messaging(ctx)
	if err != nil {
		fmt.Println("Failed to instantiate firebase cloud messaging client: ", err)
		os.Exit(1)
	}
	notificationService := service.NewFirebaseNotificationService(messagingClient, config.Firebase.DryRun)

	notificationSubscriber, err := pubsub.OpenSubscription(ctx, config.NotificationTopic.GetSqsUri())
	if err != nil {
		fmt.Println("Failed to initialize notification sqs subscription topic: ", err)
		os.Exit(1)
	}
	defer notificationSubscriber.Shutdown(ctx)
	notificationSubscriberService := service.NewSqsSubService(notificationSubscriber)

	notificationPublisher, err := pubsub.OpenTopic(ctx, config.NotificationTopic.GetSqsUri())
	if err != nil {
		fmt.Println("Failed to initialize notification sqs publisher topic: ", err)
		os.Exit(1)
	}
	defer notificationPublisher.Shutdown(ctx)
	notificationPublisherService := service.NewSqsPubService(notificationPublisher)

	taxicallSubscriber, err := pubsub.OpenSubscription(ctx, config.TaxicallTopic.GetSqsUri())
	if err != nil {
		fmt.Println("Failed to initialize taxicall sqs subscription topic: ", err)
		os.Exit(1)
	}
	defer taxicallSubscriber.Shutdown(ctx)
	taxicallSubscriberService := service.NewSqsSubService(taxicallSubscriber)

	taxicallPublisher, err := pubsub.OpenTopic(ctx, config.TaxicallTopic.GetSqsUri())
	if err != nil {
		fmt.Println("Failed to initialize taxicall sqs publisher topic: ", err)
		os.Exit(1)
	}
	defer taxicallPublisher.Shutdown(ctx)
	taxicallPublisherService := service.NewSqsPubService(taxicallPublisher)

	// Init apps
	pushApp, err := push.NewPushApp(
		push.WithTransactor(transactor),
		push.WithRouteService(mapRouteService),
		push.WithNotificationService(notificationService),
		push.WithPushTokenRepository(pushTokenRepository),
		push.WithEventSubscribeService(notificationSubscriberService),
		push.WithEventPublisherService(notificationPublisherService),
	)
	if err != nil {
		fmt.Printf("Failed to setup push app: %v\n", err)
		os.Exit(1)
	}

	if err := pushApp.Start(ctx); err != nil {
		fmt.Printf("Failed to start push app event loop: %v\n", err)
		os.Exit(1)
	}
	defer pushApp.Stop(ctx)

	taxicallApp, err := taxicall.NewTaxicallApp(
		taxicall.WithTransactor(transactor),
		taxicall.WithDriverLocationRepository(driverLocationRepository),
		taxicall.WithTaxiCallRequestRepository(taxiCallRequestRepository),
		taxicall.WithEventRepository(eventRepository),
		taxicall.WithRouteServie(mapRouteService),
		taxicall.WithLocationService(locationService),
		taxicall.WithEventPublisherService(taxicallPublisherService),
		taxicall.WithEventSubscriberService(taxicallSubscriberService),
	)
	if err != nil {
		fmt.Printf("Failed to start taxi call app: %v\n", err)
		os.Exit(1)
	}
	defer taxicallApp.Shutdown(ctx)

	notificationOutboxApp, err := outbox.NewOutboxApp(
		outbox.WithTransactor(transactor),
		outbox.WithEventRepository(eventRepository),
		outbox.WithEventPublishService(notificationPublisherService),
		outbox.WithTargetEventUirs(config.NotificationOutbox.EventUris),
		outbox.WithPollInterval(config.NotificationOutbox.PollInterval),
		outbox.WithMaxMessages(config.NotificationOutbox.MaxMessages),
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

	taxicallOutboxApp, err := outbox.NewOutboxApp(
		outbox.WithTransactor(transactor),
		outbox.WithEventRepository(eventRepository),
		outbox.WithEventPublishService(taxicallPublisherService),
		outbox.WithTargetEventUirs(config.TaxicallOutbox.EventUris),
		outbox.WithPollInterval(config.TaxicallOutbox.PollInterval),
		outbox.WithMaxMessages(config.TaxicallOutbox.MaxMessages),
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

	if err := taxicallApp.Start(ctx); err != nil {
		fmt.Printf("Failed to start taxi call app event loop: %v\n", err)
		os.Exit(1)
	}

	userSessionApp, err := usersession.NewUserSessionApp(
		usersession.WithTransactor(transactor),
		usersession.WithUserSessionRepository(userSessionRepository),
	)
	if err != nil {
		fmt.Printf("Failed to setup user session app: %v\n", err)
		os.Exit(1)
	}

	driverSessionApp, err := driversession.NewDriverSessionApp(
		driversession.WithTransactor(transactor),
		driversession.WithDriverSessionRepository(driverSessionRepository),
	)
	if err != nil {
		fmt.Printf("Failed to setup driver session app: %v\n", err)
		os.Exit(1)
	}

	userApp, err := user.NewUserApp(
		user.WithTransactor(transactor),
		user.WithUserRepository(userRepository),
		user.WithSessionService(userSessionApp),
		user.WithSmsVerificationRepository(smsVerificationRepository),
		user.WithUserPaymentRepository(userPaymentRepository),
		user.WithCardPaymentService(tossPaymentService),
		user.WithSmsSenderService(smsSenderService),
		user.WithMapRouteService(mapRouteService),
		user.WithLocationService(locationService),
		user.WithPushService(pushApp),
		user.WithTaxiCallService(taxicallApp),
	)
	if err != nil {
		fmt.Printf("Failed to setup user app: %v\n", err)
		os.Exit(1)
	}

	driverApp, err := driver.NewDriverApp(
		driver.WithTransactor(transactor),
		driver.WithDriverRepository(driverRepository),
		driver.WithSettlementAccountRepository(driverSettlementAccountRepository),
		driver.WithSessionService(driverSessionApp),
		driver.WithSmsSenderService(smsSenderService),
		driver.WithSmsVerificationRepository(smsVerificationRepository),
		driver.WithFileUploadService(fileUploadService),
		driver.WithEventRepository(eventRepository),
		driver.WithPushService(pushApp),
		driver.WithTaxiCallService(taxicallApp),
	)
	if err != nil {
		fmt.Printf("Failed to setup driver app: %v\n", err)
		os.Exit(1)
	}

	// Init middlewares
	userSessionMiddleware := userserver.NewSessionMiddleware(userSessionApp)

	driverSessionMiddleware := driverserver.NewSessionMiddleware(driverSessionApp)

	backofficeSessionMiddleware := backofficeserver.NewSessionMiddleware(config.Backoffice.Secret)

	// Init servers
	userServer, err := userserver.NewUserServer(
		userserver.WithEndpoint("0.0.0.0"),
		userserver.WithPort(18881),
		userserver.WithUserApp(userApp),
		userserver.WithMiddleware(userSessionMiddleware.Get()),
		userserver.WithMiddleware(userserver.UserIdChecker),
	)
	if err != nil {
		fmt.Printf("Failed to setup user server: %v\n", err)
		os.Exit(1)
	}
	defer userServer.Stop(ctx)

	driverServer, err := driverserver.NewDriverServer(
		driverserver.WithEndpoint("0.0.0.0"),
		driverserver.WithPort(18882),
		driverserver.WithDriverApp(driverApp),
		driverserver.WithMiddleware(driverSessionMiddleware.Get()),
		driverserver.WithMiddleware(driverserver.DriverIdChecker),
	)
	if err != nil {
		fmt.Printf("Failed to setup driver server: %v\n", err)
		os.Exit(1)
	}
	defer driverServer.Stop(ctx)

	backofficeServer, err := backofficeserver.NewBackofficeServer(
		backofficeserver.WithEndpoint("0.0.0.0"),
		backofficeserver.WithPort(18883),
		backofficeserver.WithDriverApp(driverApp),
		backofficeserver.WithUserApp(userApp),
		backofficeserver.WithMiddleware(backofficeSessionMiddleware.Get()),
	)
	if err != nil {
		fmt.Printf("Failed to setup backoffice server: %v\n", err)
		os.Exit(1)
	}
	defer backofficeServer.Stop(ctx)

	go func() {
		if err := userServer.Run(ctx); err != nil && err != http.ErrServerClosed {
			// TODO (taekyeom) fatal log
			fmt.Printf("shutting down user server:\n%v", err)
		}
	}()

	go func() {
		if err := driverServer.Run(ctx); err != nil && err != http.ErrServerClosed {
			// TODO (taekyeom) fatal log
			fmt.Printf("shutting down driver server:\n%v", err)
		}
	}()

	go func() {
		if err := backofficeServer.Run(ctx); err != nil && err != http.ErrServerClosed {
			// TODO (taekyeom) fatal log
			fmt.Printf("shutting down backoffice server:\n%v", err)
		}
	}()

	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	fmt.Println("shutting down [Taco-Backend] service... because of interrupt")
	cancel()
}
