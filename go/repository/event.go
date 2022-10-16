package repository

import (
	"context"
	"fmt"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/uptrace/bun"
)

type EventRepository interface {
	BatchGet(context.Context, bun.IDB, []string, int) ([]entity.Event, error)
	BatchCommit(context.Context, bun.IDB, []entity.Event) error
	BatchCreate(context.Context, bun.IDB, []entity.Event) error
}

type eventRepository struct{}

func (e eventRepository) BatchGet(ctx context.Context, db bun.IDB, eventUris []string, maxSize int) ([]entity.Event, error) {
	resp := []entity.Event{}

	err := db.NewSelect().Model(&resp).
		Where("event_uri IN (?)", bun.In(eventUris)).
		Order("create_time").
		Limit(maxSize).
		Scan(ctx)

	if err != nil {
		return []entity.Event{}, fmt.Errorf("%w: erorr from db: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (e eventRepository) BatchCommit(ctx context.Context, db bun.IDB, events []entity.Event) error {
	res, err := db.NewDelete().Model(&events).WherePK().Exec(ctx)

	if err != nil {
		return fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}
	if rowsAffected != int64(len(events)) {
		return fmt.Errorf("%w: invalid rows affected %d", value.ErrDBInternal, rowsAffected)
	}

	return nil
}

func (e eventRepository) BatchCreate(ctx context.Context, db bun.IDB, events []entity.Event) error {
	res, err := db.NewInsert().Model(&events).Exec(ctx)

	if err != nil {
		return fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}
	if rowsAffected != int64(len(events)) {
		return fmt.Errorf("%w: invalid rows affected %d", value.ErrDBInternal, rowsAffected)
	}

	return nil
}

func NewEventRepository() eventRepository {
	return eventRepository{}
}
