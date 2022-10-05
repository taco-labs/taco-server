package repository

import (
	"context"
	"fmt"

	"github.com/taco-labs/taco/go/domain/value"
	"github.com/uptrace/bun"
)

type Transactor interface {
	Run(context.Context, func(context.Context, bun.Tx) error) error
}

type defaultTransactor struct {
	db *bun.DB
}

func (d defaultTransactor) Run(ctx context.Context, fn func(context.Context, bun.IDB) error) error {
	qCtx := GetQueryContext(ctx)
	if qCtx != nil {
		return fn(ctx, qCtx)
	}

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%w: error while open transaction:\n%v", value.ErrDBInternal, err)
	}
	ctx = WithQueryContext(ctx, &tx)
	err = fn(ctx, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func NewDefaultTransactor(db *bun.DB) defaultTransactor {
	return defaultTransactor{db}
}
