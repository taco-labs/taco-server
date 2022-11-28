package driversession

import (
	"context"
	"errors"
	"fmt"

	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/repository"
	"github.com/uptrace/bun"
)

type driverSessionApp struct {
	app.Transactor
	repository struct {
		session repository.DriverSessionRepository
	}
}

func (d driverSessionApp) GetById(ctx context.Context, sessionId string) (entity.DriverSession, error) {
	var driverSession entity.DriverSession
	err := d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		sess, err := d.repository.session.GetById(ctx, i, sessionId)
		if err != nil {
			return fmt.Errorf("app.DriverSession.Get: error while get session:\n%w", err)
		}
		driverSession = sess
		return nil
	})

	if err != nil {
		return entity.DriverSession{}, err
	}

	return driverSession, nil
}

func (d driverSessionApp) RevokeByDriverId(ctx context.Context, driverId string) error {
	return d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		if err := d.repository.session.DeleteByDriverId(ctx, i, driverId); err != nil {
			return fmt.Errorf("app.DriverSession.RevokeByDriverId: error while revoke driver session:\n%w", err)
		}
		return nil
	})
}

func (d driverSessionApp) Create(ctx context.Context, session entity.DriverSession) error {
	return d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		if err := d.repository.session.Create(ctx, i, session); err != nil {
			return fmt.Errorf("app.DriverSession.Create: error while create driver session:\n%w", err)
		}
		return nil
	})
}

func (d driverSessionApp) Update(ctx context.Context, session entity.DriverSession) error {
	return d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		if err := d.repository.session.Update(ctx, i, session); err != nil {
			return fmt.Errorf("app.DriverSession.Update: error while create driver session:\n%w", err)
		}
		return nil
	})
}

func (d driverSessionApp) ActivateByDriverId(ctx context.Context, driverId string) error {
	return d.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		if err := d.repository.session.ActivateByDriverId(ctx, i, driverId); err != nil {
			return fmt.Errorf("app.DriverSession.Create: error while activate driver session:\n%w", err)
		}

		return nil
	})
}

func NewDriverSessionApp(opts ...driverSessionOption) (driverSessionApp, error) {
	dsa := driverSessionApp{}

	for _, opt := range opts {
		opt(&dsa)
	}

	return dsa, dsa.validate()
}

func (d driverSessionApp) validate() error {
	if d.Transactor == nil {
		return errors.New("driver session app need transactor")
	}

	if d.repository.session == nil {
		return errors.New("driver session app need session repository")
	}

	return nil
}
