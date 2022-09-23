package driversession

import (
	"context"
	"errors"
	"fmt"

	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/repository"
)

type driverSessionApp struct {
	app.Transactor
	repository struct {
		session repository.DriverSessionRepository
	}
}

func (d driverSessionApp) GetById(ctx context.Context, sessionId string) (entity.DriverSession, error) {
	ctx, err := d.Start(ctx)
	if err != nil {
		return entity.DriverSession{}, err
	}

	defer func() {
		err = d.Done(ctx, err)
	}()

	driverSession, err := d.repository.session.GetById(ctx, sessionId)
	if err != nil {
		return entity.DriverSession{}, fmt.Errorf("app.DriverSession.Get: error while get session:\n%w", err)
	}

	return driverSession, nil
}

func (d driverSessionApp) RevokeByDriverId(ctx context.Context, driverId string) error {
	ctx, err := d.Start(ctx)
	if err != nil {
		return err
	}

	defer func() {
		err = d.Done(ctx, err)
	}()

	if err = d.repository.session.DeleteByDriverId(ctx, driverId); err != nil {
		return fmt.Errorf("app.DriverSession.RevokeByDriverId: error while revoke driver session:\n%w", err)
	}

	return nil
}

func (d driverSessionApp) Create(ctx context.Context, session entity.DriverSession) error {
	ctx, err := d.Start(ctx)
	if err != nil {
		return err
	}

	defer func() {
		err = d.Done(ctx, err)
	}()

	if err = d.repository.session.Create(ctx, session); err != nil {
		return fmt.Errorf("app.DriverSession.Create: error while create driver session:\n%w", err)
	}

	return nil
}

func (d driverSessionApp) ActivateByDriverId(ctx context.Context, driverId string) error {
	ctx, err := d.Start(ctx)
	if err != nil {
		return err
	}

	defer func() {
		err = d.Done(ctx, err)
	}()

	if err = d.repository.session.ActivateByDriverId(ctx, driverId); err != nil {
		return fmt.Errorf("app.DriverSession.Create: error while activate driver session:\n%w", err)
	}

	return nil
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
