package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"

	"github.com/ktk1012/taco/go/app"
	"github.com/ktk1012/taco/go/app/session"
	"github.com/ktk1012/taco/go/app/user"
	"github.com/ktk1012/taco/go/repository"
	"github.com/ktk1012/taco/go/server"
	"github.com/ktk1012/taco/go/service"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
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

	userRepository := repository.NewUserRepository()
	userSessionRepository := repository.NewUserSessionRepository()

	// TODO(taekyeom) Replace mock to real one
	mockIdentityService := service.NewMockIdentityService()

	userSessionApp, err := session.NewUserSessionApp(
		session.WithTransactor(transactor),
		session.WithUserSessionRepository(userSessionRepository),
	)

	if err != nil {
		fmt.Printf("Failed to setup user session app: %v\n", err)
		os.Exit(1)
	}

	userApp, err := user.NewUserApp(
		user.WithTransactor(transactor),
		user.WithUserRepository(userRepository),
		user.WithSessionService(userSessionApp),
		user.WithUserIdentityService(mockIdentityService),
	)

	if err != nil {
		fmt.Printf("Failed to setup user app: %v\n", err)
		os.Exit(1)
	}

	userServer, err := server.NewUserServer(
		"0.0.0.0",
		18881,
		userApp,
	)

	if err != nil {
		fmt.Printf("Failed to setup user server: %v\n", err)
		os.Exit(1)
	}

	if err = userServer.Run(ctx); err != nil {
		fmt.Printf("Failed to start user server: %v\n", err)
		os.Exit(1)
	}
}
