package session

import (
	"fmt"

	"github.com/ktk1012/taco/go/app"
	"github.com/ktk1012/taco/go/domain/entity"
	"github.com/ktk1012/taco/go/repository"
	"golang.org/x/net/context"
)

type userSessionApp struct {
	app.Transactor
	repository struct {
		session repository.UserSessionRepository
	}
}

func (u userSessionApp) GetSession(ctx context.Context, sessionId string) (entity.UserSession, error) {
	ctx, err := u.Start(ctx)
	if err != nil {
		return entity.UserSession{}, err
	}

	defer func() {
		err = u.Done(ctx, err)
	}()

	userSession, err := u.repository.session.GetSession(ctx, sessionId)
	if err != nil {
		return entity.UserSession{}, fmt.Errorf("app.UserSession.GetSession: error while get session:\n %v", err)
	}

	return userSession, nil
}

func (u userSessionApp) RevokeSessionByUserId(ctx context.Context, userId string) error {
	ctx, err := u.Start(ctx)
	if err != nil {
		return err
	}

	defer func() {
		err = u.Done(ctx, err)
	}()

	if err = u.repository.session.DeleteSessionByUserId(ctx, userId); err != nil {
		return fmt.Errorf("app.UserSession.DeleteSessionByUserId: error while delete session:\n %v", err)
	}

	return nil
}

func (u userSessionApp) CreateSession(ctx context.Context, session entity.UserSession) error {
	ctx, err := u.Start(ctx)
	if err != nil {
		return err
	}

	defer func() {
		err = u.Done(ctx, err)
	}()

	if err = u.repository.session.CreateSession(ctx, session); err != nil {
		return fmt.Errorf("app.UserSession.CreateSession: error while create session:\n %v", err)
	}

	return nil
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
