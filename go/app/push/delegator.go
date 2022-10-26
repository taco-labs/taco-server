package push

import (
	"context"
	"fmt"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
)

type userGetterDelegator struct {
	delegatee userGetterInterface
}

func (u *userGetterDelegator) Set(delegatee userGetterInterface) {
	u.delegatee = delegatee
}

func (u userGetterDelegator) GetUser(ctx context.Context, userId string) (entity.User, error) {
	if u.delegatee == nil {
		return entity.User{}, fmt.Errorf("%w: No user getter service delegatee", value.ErrInternal)
	}
	return u.delegatee.GetUser(ctx, userId)
}

func NewUserGetterDelegator() *userGetterDelegator {
	return &userGetterDelegator{}
}

type driverGetterDelegator struct {
	delegatee driverGetterInterface
}

func (u *driverGetterDelegator) Set(delegatee driverGetterInterface) {
	u.delegatee = delegatee
}

func (u driverGetterDelegator) GetDriver(ctx context.Context, driverId string) (entity.Driver, error) {
	if u.delegatee == nil {
		return entity.Driver{}, fmt.Errorf("%w: No driver getter service delegatee", value.ErrInternal)
	}
	return u.delegatee.GetDriver(ctx, driverId)
}

func NewDriverGetterDelegator() *driverGetterDelegator {
	return &driverGetterDelegator{}
}
