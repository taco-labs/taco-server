package user

type userAppDelegator struct {
	*userApp
}

func (u *userAppDelegator) Set(delegatee *userApp) {
	u.userApp = delegatee
}

func NewUserAppDelegator() *userAppDelegator {
	return &userAppDelegator{}
}
