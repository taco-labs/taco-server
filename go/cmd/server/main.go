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
)

func main() {
	ctx := context.Background()

	serverConfig, err := config.NewServerConfig(ctx)
	if err != nil {
		fmt.Printf("Failed to initialize taco config: %v\n", err)
		os.Exit(1)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("Failed to initializae logger: %v\n", err)
		os.Exit(1)
	}

	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	err = cmd.RunServer(ctx, serverConfig, logger, quit)
	if err != nil {
		fmt.Printf("Error while start server: %v\n", err)
		os.Exit(1)
	}
}
