package utils

import (
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

var Logger *zap.Logger

type loggerKey struct{}

func init() {
	Logger, _ = zap.NewProduction()
}

func GetLogger(ctx context.Context) *zap.Logger {
	logger := ctx.Value(loggerKey{})

	return logger.(*zap.Logger)
}

func SetLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}
