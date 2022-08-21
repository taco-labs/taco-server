package repository

import (
	"context"

	"github.com/uptrace/bun"
)

type queryContextKey struct{}

func WithQueryContext(ctx context.Context, q bun.IDB) context.Context {
	return context.WithValue(ctx, queryContextKey{}, q)
}

func GetQueryContext(ctx context.Context) bun.IDB {
	v, ok := ctx.Value(queryContextKey{}).(bun.IDB)
	if !ok {
		return nil
	}
	return v
}
