package usersession

import (
	"fmt"

	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/repository"
	"github.com/uptrace/bun"
	"golang.org/x/net/context"
)

type userSessionApp struct {
	app.Transactor
	repository struct {
		session repository.UserSessionRepository
	}
}

func (u userSessionApp) GetSession(ctx context.Context, sessionId string) (entity.UserSession, error) {
	var userSession entity.UserSession
	var err error

	err = u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		userSession, err = u.repository.session.GetById(ctx, i, sessionId)
		if err != nil {
			return fmt.Errorf("app.UserSession.GetSession: error while get session: %w", err)
		}
		return nil
	})

	if err != nil {
		return entity.UserSession{}, err
	}

	return userSession, nil
}

func (u userSessionApp) RevokeSessionByUserId(ctx context.Context, userId string) error {
	return u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		if err := u.repository.session.DeleteByUserId(ctx, i, userId); err != nil {
			return fmt.Errorf("app.UserSession.DeleteSessionByUserId: error while delete session: %w", err)
		}
		return nil
	})
}

func (u userSessionApp) CreateSession(ctx context.Context, session entity.UserSession) error {
	return u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		if err := u.repository.session.Create(ctx, i, session); err != nil {
			return fmt.Errorf("app.UserSession.CreateSession: error while create session: %w", err)
		}

		return nil
	})
}

func (u userSessionApp) UpdateSession(ctx context.Context, session entity.UserSession) error {
	return u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		if err := u.repository.session.Update(ctx, i, session); err != nil {
			return fmt.Errorf("app.UserSession.UpdateSession: error while update session: %w", err)
		}

		return nil
	})
}

func (u userSessionApp) validateApp() error {
	if u.Transactor == nil {
		return fmt.Errorf("userSessionApp need Transactor")
	}

	if u.repository.session == nil {
		return fmt.Errorf("userSessionApp need session")
	}

	return nil
}

func NewUserSessionApp(opts ...userSessionOption) (userSessionApp, error) {
	userSessionApp := userSessionApp{}

	for _, opt := range opts {
		opt(&userSessionApp)
	}

	return userSessionApp, userSessionApp.validateApp()
}
