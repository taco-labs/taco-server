package push

import (
	"context"
	"errors"
	"fmt"

	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/repository"
	"github.com/taco-labs/taco/go/service"
	"github.com/taco-labs/taco/go/utils"
	"github.com/uptrace/bun"
)

type userGetterInterface interface {
	GetUser(context.Context, string) (entity.User, error)
}

type driverGetterInterface interface {
	GetDriver(context.Context, string) (entity.Driver, error)
}

type taxiCallPushApp struct {
	app.Transactor
	repository struct {
		pushToken repository.PushTokenRepository
	}
	service struct {
		route        service.MapRouteService
		notification service.NotificationService
		eventSub     service.EventSubscriptionService
		userGetter   userGetterInterface
		driverGetter driverGetterInterface
		workerPool   service.WorkerPoolService
	}
	waitCh chan struct{}
}

func (t taxiCallPushApp) CreatePushToken(ctx context.Context, req request.CreatePushTokenRequest) (entity.PushToken, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)
	pushToken := entity.PushToken{
		PrincipalId: req.PrincipalId,
		FcmToken:    req.FcmToken,
		CreateTime:  requestTime,
		UpdateTime:  requestTime,
	}
	err := t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		pt, err := t.repository.pushToken.Get(ctx, i, pushToken.PrincipalId)
		if err != nil && !errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("app.push.CreatePushToken: error while get push token: %w", err)
		}
		if pt.PrincipalId != "" {
			return fmt.Errorf("app.push.CreatePustToken: push token already exist: %w", value.ErrAlreadyExists)
		}

		if err := t.repository.pushToken.Create(ctx, i, pushToken); err != nil {
			return fmt.Errorf("app.push.CreatePushToken: error while create push token: %w", err)
		}

		return nil
	})

	if err != nil {
		return entity.PushToken{}, err
	}

	return pushToken, nil
}

func (t taxiCallPushApp) UpdatePushToken(ctx context.Context, req request.UpdatePushTokenRequest) error {
	requestTime := utils.GetRequestTimeOrNow(ctx)
	return t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		pt, err := t.repository.pushToken.Get(ctx, i, req.PrincipalId)
		if err != nil && errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("app.push.UpdatePushToken: push token not found: %w", err)
		}
		if err != nil {
			return fmt.Errorf("app.push.UpdatePushToken: error while get push token: %w", err)
		}

		pt.FcmToken = req.FcmToken
		pt.UpdateTime = requestTime

		if err := t.repository.pushToken.Update(ctx, i, pt); err != nil {
			return fmt.Errorf("app.push.UpdatePushToken: error while update push token: %w", err)
		}

		return nil
	})
}

func (t taxiCallPushApp) DeletePushToken(ctx context.Context, principalId string) error {
	return t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		pt, err := t.repository.pushToken.Get(ctx, i, principalId)
		if err != nil && errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("app.push.DeletePushToken: push token not found: %w", err)
		}
		if err != nil {
			return fmt.Errorf("app.push.DeletePushToken: error while get push token: %w", err)
		}

		if err := t.repository.pushToken.Delete(ctx, i, pt); err != nil {
			return fmt.Errorf("app.push.DeletePushToken: error while delete push token: %w", err)
		}

		return nil
	})
}
