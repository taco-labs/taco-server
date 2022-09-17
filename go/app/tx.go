package app

import (
	"context"
	"fmt"

	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/repository"
	"github.com/uptrace/bun"
)

type transactorCurrentScopeKey struct{}

var (
	nilFunc = func() error { return nil }
)

type Transactor interface {
	Start(context.Context) (context.Context, error)
	Done(context.Context, error) error
}

type defaultTransactor struct {
	db *bun.DB
}

func (t defaultTransactor) Start(ctx context.Context) (context.Context, error) {
	qCtx := repository.GetQueryContext(ctx)
	if qCtx != nil {
		return context.WithValue(ctx, transactorCurrentScopeKey{}, false), nil
	}

	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return ctx, fmt.Errorf("%v: error while open transaction", value.ErrDBInternal)
	}

	ctx = repository.WithQueryContext(ctx, &tx)
	ctx = context.WithValue(ctx, transactorCurrentScopeKey{}, true)

	return ctx, nil
}

func (t defaultTransactor) Done(ctx context.Context, err error) error {
	tx, ok := repository.GetQueryContext(ctx).(*bun.Tx)
	if !ok {
		return err
	}

	isCurrentScope := ctx.Value(transactorCurrentScopeKey{}).(bool)
	if !isCurrentScope {
		return nil
	}

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func NewDefaultTranscator(db *bun.DB) defaultTransactor {
	return defaultTransactor{db}
}
