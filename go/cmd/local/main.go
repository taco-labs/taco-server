package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/taco-labs/taco/go/cmd"
	"github.com/taco-labs/taco/go/config"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx := context.Background()

	serverConfig, err := config.NewServerConfig(ctx)
	if err != nil {
		fmt.Printf("Failed to initialize taco config: %v\n", err)
		os.Exit(1)
	}

	outboxConfig, err := config.NewOutboxConfig(ctx)
	if err != nil {
		fmt.Printf("Failed to initialize outbox config: %v\n", err)
		os.Exit(1)
	}

	var logger *zap.Logger

	if serverConfig.Env == "prod" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		fmt.Printf("Failed to initializae logger: %v\n", err)
		os.Exit(1)
	}

	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quitServer := make(chan os.Signal, 1)
	signal.Notify(quitServer, os.Interrupt, syscall.SIGTERM)

	quitOutbox := make(chan os.Signal, 1)
	signal.Notify(quitOutbox, os.Interrupt, syscall.SIGTERM)

	errgroup, gCtx := errgroup.WithContext(ctx)

	errgroup.Go(func() error {
		return cmd.RunServer(gCtx, serverConfig, logger, quitServer)
	})

	errgroup.Go(func() error {
		return cmd.RunOutbox(gCtx, outboxConfig, logger, quitOutbox)
	})

	if err := errgroup.Wait(); err != nil {
		fmt.Printf("Error while start services: %v\n", err)
		os.Exit(1)
	}
}
