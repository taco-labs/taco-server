package utils

import (
	"context"
	"time"
)

type requestTimeKey struct{}
type userIdKey struct{}

func SetRequestTime(ctx context.Context, requestTime time.Time) context.Context {
	return context.WithValue(ctx, requestTimeKey{}, requestTime)
}

func GetRequestTime(ctx context.Context) time.Time {
	return ctx.Value(requestTimeKey{}).(time.Time)
}

func GetRequestTimeOrNow(ctx context.Context) time.Time {
	v, ok := ctx.Value(requestTimeKey{}).(time.Time)
	if !ok {
		return time.Now()
	}
	return v
}

func SetUserId(ctx context.Context, userId string) context.Context {
	return context.WithValue(ctx, userIdKey{}, userId)
}

func GetUserId(ctx context.Context) string {
	return ctx.Value(userIdKey{}).(string)
}
