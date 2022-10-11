package app

import (
	"context"
	"errors"
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
	Run(context.Context, func(context.Context, bun.IDB) error) error
	RunWithNonRollbackError(context.Context, error, func(context.Context, bun.IDB) error) error
}

type defaultTransactor struct {
	db *bun.DB
}

func (d defaultTransactor) Run(ctx context.Context, fn func(context.Context, bun.IDB) error) error {
	qCtx := repository.GetQueryContext(ctx)
	if qCtx != nil {
		return fn(ctx, qCtx)
	}

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%w: error while open transaction:\n%v", value.ErrDBInternal, err)
	}

	ctx = repository.WithQueryContext(ctx, &tx)
	err = fn(ctx, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (d defaultTransactor) RunWithNonRollbackError(ctx context.Context, nonRollbackError error, fn func(context.Context, bun.IDB) error) error {
	qCtx := repository.GetQueryContext(ctx)
	if qCtx != nil {
		return fn(ctx, qCtx)
	}

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%w: error while open transaction:\n%v", value.ErrDBInternal, err)
	}
	ctx = repository.WithQueryContext(ctx, &tx)
	err = fn(ctx, tx)
	if err != nil && !errors.Is(err, nonRollbackError) {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("Error while commit: %v, %w", err, value.ErrDBInternal)
	}

	return err
}

func NewDefaultTranscator(db *bun.DB) defaultTransactor {
	return defaultTransactor{db}
}
