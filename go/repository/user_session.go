package repository

import (
	"context"
	"errors"

	"github.com/taco-labs/taco/go/domain/entity"
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

	// TODO (taekyeom) error handling
	if err != nil {
		return entity.UserSession{}, err
	}

	return userSession, nil
}

func (u userSessionRepository) DeleteSessionByUserId(ctx context.Context, db bun.IDB, userId string) error {
	userSession := entity.UserSession{}

	res, err := db.NewDelete().Model(&userSession).Where("user_id = ?", userId).Exec(ctx)

	// TODO(taekyeom) Error handling
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	// TODO(taekyeom) Error handling
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		// TODO(taekyeom) Error handling
		return errors.New("invalid update")
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
	// TODO(taekyeom) Error handling
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		// TODO(taekyeom) Error handling
		return errors.New("invalid update")
	}

	return nil
}

func NewUserSessionRepository() userSessionRepository {
	return userSessionRepository{}
}
