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

type UserSessionRepository interface {
	GetSession(context.Context, bun.IDB, string) (entity.UserSession, error)
	CreateSession(context.Context, bun.IDB, entity.UserSession) error
	DeleteSessionByUserId(context.Context, bun.IDB, string) error
}

type userSessionRepository struct{}

func (u userSessionRepository) GetSession(ctx context.Context, db bun.IDB, sessionId string) (entity.UserSession, error) {
	userSession := entity.UserSession{Id: sessionId}

	err := db.NewSelect().Model(&userSession).WherePK().Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.UserSession{}, value.ErrNotFound
	}
	if err != nil {
		return entity.UserSession{}, fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return userSession, nil
}

func (u userSessionRepository) DeleteSessionByUserId(ctx context.Context, db bun.IDB, userId string) error {
	userSession := entity.UserSession{}

	res, err := db.NewDelete().Model(&userSession).Where("user_id = ?", userId).Exec(ctx)

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

func (u userSessionRepository) CreateSession(ctx context.Context, db bun.IDB, userSession entity.UserSession) error {
	res, err := db.NewInsert().Model(&userSession).Exec(ctx)

	// TODO(taekyeom) Error handling
	if err != nil {
		return err
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

func NewUserSessionRepository() *userSessionRepository {
	return &userSessionRepository{}
}
