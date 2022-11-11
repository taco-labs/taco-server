package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/uptrace/bun"
)

type PushTokenRepository interface {
	Get(context.Context, bun.IDB, string) (entity.PushToken, error)
	Update(context.Context, bun.IDB, entity.PushToken) error
	Create(context.Context, bun.IDB, entity.PushToken) error
	Delete(context.Context, bun.IDB, entity.PushToken) error
}

type pushTokenRepository struct{}

func (p pushTokenRepository) Get(ctx context.Context, db bun.IDB, principalId string) (entity.PushToken, error) {
	resp := entity.PushToken{
		PrincipalId: principalId,
	}

	err := db.NewSelect().Model(&resp).WherePK().Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return entity.PushToken{}, value.ErrNotFound
	}
	if err != nil {
		return entity.PushToken{}, fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (p pushTokenRepository) Update(ctx context.Context, db bun.IDB, pushToken entity.PushToken) error {
	res, err := db.NewUpdate().Model(&pushToken).WherePK().Exec(ctx)

	if err != nil {
		return fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}
	if rowsAffected != 1 {
		return fmt.Errorf("%w: invalid rows affected %d", value.ErrDBInternal, rowsAffected)
	}

	return nil
}

func (p pushTokenRepository) Create(ctx context.Context, db bun.IDB, pushToken entity.PushToken) error {
	res, err := db.NewInsert().Model(&pushToken).Exec(ctx)

	if err != nil {
		return fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}
	if rowsAffected != 1 {
		return fmt.Errorf("%w: invalid rows affected %d", value.ErrDBInternal, rowsAffected)
	}

	return nil
}

func (p pushTokenRepository) Delete(ctx context.Context, db bun.IDB, pushToken entity.PushToken) error {
	res, err := db.NewDelete().Model(&pushToken).WherePK().Exec(ctx)

	if err != nil {
		return fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}
	if rowsAffected != 1 {
		return fmt.Errorf("%w: invalid rows affected %d", value.ErrDBInternal, rowsAffected)
	}

	return nil
}

func NewPushTokenRepository() *pushTokenRepository {
	return &pushTokenRepository{}
}
