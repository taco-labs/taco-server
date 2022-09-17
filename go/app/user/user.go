package user

import (
	"errors"
	"fmt"
	"time"

	"context"

	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/domain/value/enum"
	"github.com/taco-labs/taco/go/repository"
	"github.com/taco-labs/taco/go/service"
	"github.com/taco-labs/taco/go/utils"
)

type SessionInterface interface {
	RevokeSessionByUserId(context.Context, string) error
	CreateSession(context.Context, entity.UserSession) error
}

type userApp struct {
	app.Transactor
	repository struct {
		user    repository.UserRepository
		payment repository.UserPaymentRepository
	}

	service struct {
		userIdentity service.UserIdentityService
		session      SessionInterface
		payment      service.CardPaymentService
	}

	actor struct {
	}
}

func (u userApp) Signup(ctx context.Context, req request.UserSignupRequest) (entity.User, string, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)
	userIdentity, err := u.service.userIdentity.GetUserIdentity(ctx, req.IamUid)
	if err != nil {
		return entity.User{}, "", err
	}

	ctx, err = u.Start(ctx)
	if err != nil {
		return entity.User{}, "", err
	}
	defer func() {
		err = u.Done(ctx, err)
	}()

	user, err := u.repository.user.FindByUserUniqueKey(ctx, userIdentity.UserUniqueKey)
	if !errors.Is(value.ErrUserNotFound, err) && err != nil {
		return entity.User{}, "", fmt.Errorf("app.User.Signup: error while find user by unique key:\n %v", err)
	}

	if errors.Is(value.ErrUserNotFound, err) {
		// create user
		newUser := entity.User{
			Id:            utils.MustNewUUID(),
			FirstName:     req.FirstName,
			LastName:      req.LastName,
			Email:         req.Email,
			BirthDay:      userIdentity.BirthDay,
			Phone:         userIdentity.Phone,
			Gender:        userIdentity.Gender,
			AppOs:         enum.OsTypeFromString(req.AppOs),
			AppVersion:    req.AppVersion,
			AppFcmToken:   req.AppFcmToken,
			UserUniqueKey: userIdentity.UserUniqueKey,
			CreateTime:    requestTime,
			UpdateTime:    requestTime,
			DeleteTime:    time.Time{},
		}
		if err = u.repository.user.CreateUser(ctx, newUser); err != nil {
			return entity.User{}, "", fmt.Errorf("app.User.Signup: error while create user:\n %v", err)
		}

		userSession := entity.UserSession{
			Id:         utils.MustNewUUID(),
			UserId:     newUser.Id,
			ExpireTime: requestTime.AddDate(0, 1, 0), // TODO(taekyeom) Configurable expire time
		}
		if err = u.service.session.CreateSession(ctx, userSession); err != nil {
			return entity.User{}, "", fmt.Errorf("app.User.Signup: error while create new session:\n %v", err)
		}

		return user, userSession.Id, nil
	} else {
		// update user
		// TODO(taekyeom) 모든 변경 사항은 method로 묶고 테스트 가능하도록 해야 함
		user.Email = req.Email
		user.AppOs = enum.OsTypeFromString(req.AppOs)
		user.AppVersion = req.AppVersion
		user.AppFcmToken = req.AppFcmToken
		user.Phone = userIdentity.Phone

		user.UpdateTime = requestTime

		if err = u.repository.user.UpdateUser(ctx, user); err != nil {
			return entity.User{}, "", fmt.Errorf("app.User.Signup: error while update user:\n %v", err)
		}

		// revoke session
		if err = u.service.session.RevokeSessionByUserId(ctx, user.Id); err != nil {
			return entity.User{}, "", fmt.Errorf("app.User.Signup: error while revoke previous session:\n %v", err)
		}

		// create new session
		userSession := entity.UserSession{
			Id:         utils.MustNewUUID(),
			UserId:     user.Id,
			ExpireTime: requestTime.AddDate(0, 1, 0), // TODO(taekyeom) Configurable expire time
		}
		if err = u.service.session.CreateSession(ctx, userSession); err != nil {
			return entity.User{}, "", fmt.Errorf("app.User.Signup: error while create new session:\n %v", err)
		}

		return user, userSession.Id, nil
	}
}

func (u userApp) UpdateUser(ctx context.Context, req request.UserUpdateRequest) (entity.User, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	ctx, err := u.Start(ctx)
	if err != nil {
		return entity.User{}, err
	}
	defer func() {
		err = u.Done(ctx, err)
	}()

	user, err := u.repository.user.FindById(ctx, req.Id)
	if err != nil {
		return entity.User{}, fmt.Errorf("app.user.UpdateUser: error while find user by id:\n %v", err)
	}

	user.AppFcmToken = req.AppFcmToken
	user.UpdateTime = requestTime

	if err = u.repository.user.UpdateUser(ctx, user); err != nil {
		return entity.User{}, fmt.Errorf("app.user.UpdateUser: error while update user:\n %v", err)
	}

	return user, nil
}

func (u userApp) GetUser(ctx context.Context, userId string) (entity.User, error) {
	ctx, err := u.Start(ctx)
	if err != nil {
		return entity.User{}, err
	}
	defer func() {
		err = u.Done(ctx, err)
	}()

	user, err := u.repository.user.FindById(ctx, userId)
	if err != nil {
		return entity.User{}, fmt.Errorf("app.user.GetUser: error while find user by id:\n %v", err)
	}

	return user, nil
}

func NewUserApp(opts ...userAppOption) (userApp, error) {
	ua := userApp{}

	for _, opt := range opts {
		opt(&ua)
	}

	if err := ua.validateApp(); err != nil {
		return userApp{}, err
	}

	return ua, nil
}

func (u userApp) validateApp() error {
	if u.Transactor == nil {
		return errors.New("user app need transator")
	}

	if u.repository.user == nil {
		return errors.New("user app need user repository")
	}

	if u.repository.payment == nil {
		return errors.New("user app need user payment repository")
	}

	if u.service.session == nil {
		return errors.New("user app need user session repository")
	}

	if u.service.userIdentity == nil {
		return errors.New("user app need user identity service")
	}

	if u.service.payment == nil {
		return errors.New("user app need card payment service")
	}

	return nil
}
