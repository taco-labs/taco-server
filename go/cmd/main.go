package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/taco-labs/taco/go/actor/taxicall"
	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/app/driver"
	"github.com/taco-labs/taco/go/app/driversession"
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
	"golang.org/x/sync/errgroup"
)

func main() {
	config, err := config.NewTacoConfig()
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

	ctx := context.Background()

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

	db := bun.NewDB(sqldb, pgdialect.New())

	if config.Log.Query {
		hook := bundebug.NewQueryHook(bundebug.WithVerbose(true))
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

	taxiCallRequestActorService, err := taxicall.NewTaxiCallActorService(
		taxicall.WithTransactor(transactor),
		taxicall.WithUserRepository(userRepository),
		taxicall.WithDriverRepository(driverRepository),
		taxicall.WithTaxiCallRequestRepository(taxiCallRequestRepository),
	)
	if err != nil {
		fmt.Printf("Failed to setup taxi call request actor service: %v\n", err)
		os.Exit(1)
	}

	if err := taxiCallRequestActorService.Init(ctx); err != nil {
		fmt.Printf("Failed to init actor system: %v\n", err)
		os.Exit(1)
	}

	// Init apps

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
		user.WithTaxiCallRequestRepository(taxiCallRequestRepository),
		user.WithMapRouteService(mapRouteService),
		user.WithLocationService(locationService),
		user.WithTaxiCallRequestActorService(taxiCallRequestActorService),
	)
	if err != nil {
		fmt.Printf("Failed to setup user app: %v\n", err)
		os.Exit(1)
	}

	driverApp, err := driver.NewDriverApp(
		driver.WithTransactor(transactor),
		driver.WithDriverRepository(driverRepository),
		driver.WithDriverLocationRepository(driverLocationRepository),
		driver.WithSettlementAccountRepository(driverSettlementAccountRepository),
		driver.WithSessionService(driverSessionApp),
		driver.WithSmsSenderService(smsSenderService),
		driver.WithSmsVerificationRepository(smsVerificationRepository),
		driver.WithFileUploadService(fileUploadService),
		driver.WithTaxiCallRequestRepository(taxiCallRequestRepository),
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

	// Run servers
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		if err = userServer.Run(ctx); err != nil {
			return fmt.Errorf("failed to start user server:\n%v", err)
		}
		return nil
	})

	group.Go(func() error {
		if err = driverServer.Run(ctx); err != nil {
			return fmt.Errorf("failed to start driver server:\n%v", err)
		}
		return nil
	})

	group.Go(func() error {
		if err = backofficeServer.Run(ctx); err != nil {
			return fmt.Errorf("failed to start backoffice server:\n%v", err)
		}
		return nil
	})

	if err := group.Wait(); err != nil {
		fmt.Printf("Failed to start services: %v\n", err)
		os.Exit(1)
	}
}
