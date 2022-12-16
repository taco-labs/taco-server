package repository

import (
	"context"
	"fmt"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/uptrace/bun"
)

type AnalyticsRepository interface {
	Create(context.Context, bun.IDB, entity.Analytics) error
	BatchCreate(context.Context, bun.IDB, []entity.Analytics) error
}

type analyticsRepository struct{}

func (a analyticsRepository) Create(ctx context.Context, db bun.IDB, analytics entity.Analytics) error {
	res, err := db.NewInsert().Model(&analytics).Exec(ctx)

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

func (a analyticsRepository) BatchCreate(ctx context.Context, db bun.IDB, analytics []entity.Analytics) error {
	if len(analytics) == 0 {
		return nil
	}

	res, err := db.NewInsert().Model(&analytics).Exec(ctx)

	if err != nil {
		return fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}
	if rowsAffected != int64(len(analytics)) {
		return fmt.Errorf("%w: invalid rows affected %d", value.ErrDBInternal, rowsAffected)
	}

	return nil
}

func NewAnalyticsRepository() *analyticsRepository {
	return &analyticsRepository{}
}
