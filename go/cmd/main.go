package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"

	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/app/driver"
	"github.com/taco-labs/taco/go/app/driversession"
	"github.com/taco-labs/taco/go/app/user"
	"github.com/taco-labs/taco/go/app/usersession"
	"github.com/taco-labs/taco/go/repository"
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
	// TODO (taekyeom) Config
	dsn := "postgres://postgres:postgres@localhost:25432/taco?sslmode=disable&search_path=taco"
	dbDsnPtr := flag.String("dsn", dsn, "dsn of database")
	flag.Parse()
	ctx := context.Background()

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(*dbDsnPtr)))

	db := bun.NewDB(sqldb, pgdialect.New())
	bundebug.NewQueryHook(bundebug.WithVerbose(true))

	transactor := app.NewDefaultTranscator(db)

	smsVerificationRepository := repository.NewSmsVerificationRepository()

	userRepository := repository.NewUserRepository()
	userSessionRepository := repository.NewUserSessionRepository()
	userPaymentRepository := repository.NewUserPaymentRepository()

	taxiCallRequestRepository := repository.NewTaxiCallRepository()

	driverRepository := repository.NewDriverRepository()
	driverLocationRepository := repository.NewDriverLocationRepository()
	driverSettlementAccountRepository := repository.NewDriverSettlementAccountRepository()
	driverSessionRepository := repository.NewDriverSessionRepository()

	smsSenderService := service.NewCoolSmsSenderService(
		"api.coolsms.co.kr",
		"01083047880",
		"NCSCVKFUSA8TPSED",
		"L25KAYEICPWCPHTIXMKLTEAKWLFFGIHQ",
	)

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

	userSessionMiddleware := userserver.NewSessionMiddleware(userSessionApp)

	// TODO(taekyeom) Replace mock to real one
	tossPaymentService := service.NewTossPaymentService(
		"https://api.tosspayments.com",
		"dGVzdF9za196WExrS0V5cE5BcldtbzUwblgzbG1lYXhZRzVSOg==",
	)

	userApp, err := user.NewUserApp(
		user.WithTransactor(transactor),
		user.WithUserRepository(userRepository),
		user.WithSessionService(userSessionApp),
		user.WithSmsVerificationRepository(smsVerificationRepository),
		user.WithUserPaymentRepository(userPaymentRepository),
		user.WithCardPaymentService(tossPaymentService),
		user.WithSmsSenderService(smsSenderService),
		user.WithTaxiCallRequestRepository(taxiCallRequestRepository),
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
	)
	if err != nil {
		fmt.Printf("Failed to setup driver app: %v\n", err)
		os.Exit(1)
	}

	userServer, err := userserver.NewUserServer(
		userserver.WithEndpoint("0.0.0.0"),
		userserver.WithPort(18881),
		userserver.WithUserApp(userApp),
		userserver.WithMiddleware(userSessionMiddleware.Get()),
	)
	if err != nil {
		fmt.Printf("Failed to setup user server: %v\n", err)
		os.Exit(1)
	}

	driverServer, err := driverserver.NewDriverServer(
		driverserver.WithEndpoint("0.0.0.0"),
		driverserver.WithPort(18882),
		driverserver.WithDriverApp(driverApp),
	)
	if err != nil {
		fmt.Printf("Failed to setup driver server: %v\n", err)
		os.Exit(1)
	}

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

	if err := group.Wait(); err != nil {
		fmt.Printf("Failed to start services: %v\n", err)
		os.Exit(1)
	}
}
