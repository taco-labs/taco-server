package repository

import (
	"context"
	"errors"

	"github.com/ktk1012/taco/go/domain/entity"
)

type UserSessionRepository interface {
	GetSession(context.Context, string) (entity.UserSession, error)
	CreateSession(context.Context, entity.UserSession) error
	DeleteSessionByUserId(context.Context, string) error
}

type userSessionRepository struct{}

func (u userSessionRepository) GetSession(ctx context.Context, sessionId string) (entity.UserSession, error) {
	db := GetQueryContext(ctx)

	userSession := entity.UserSession{Id: sessionId}

	err := db.NewSelect().Model(&userSession).WherePK().Scan(ctx)

	// TODO (taekyeom) error handling
	if err != nil {
		return entity.UserSession{}, err
	}

	return userSession, nil
}

func (u userSessionRepository) DeleteSessionByUserId(ctx context.Context, userId string) error {
	db := GetQueryContext(ctx)

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

func (u userSessionRepository) CreateSession(ctx context.Context, userSession entity.UserSession) error {
	db := GetQueryContext(ctx)

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
