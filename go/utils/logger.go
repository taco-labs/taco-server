package utils

import (
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type loggerKey struct{}

func GetLogger(ctx context.Context) *zap.Logger {
	logger := ctx.Value(loggerKey{})

	return logger.(*zap.Logger)
}

func SetLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}
