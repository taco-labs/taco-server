package driver

type driverAppDelegator struct {
	*driverApp
}

func (u *driverAppDelegator) Set(delegatee *driverApp) {
	u.driverApp = delegatee
}

func NewDriverAppDelegator() *driverAppDelegator {
	return &driverAppDelegator{}
}
