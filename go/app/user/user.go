package user

import (
	"errors"
	"fmt"
	"time"

	"context"

	"github.com/taco-labs/taco/go/actor/taxicall"
	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/request"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/domain/value/enum"
	"github.com/taco-labs/taco/go/repository"
	"github.com/taco-labs/taco/go/service"
	"github.com/taco-labs/taco/go/utils"
	"github.com/uptrace/bun"
)

type sessionServiceInterface interface {
	RevokeSessionByUserId(context.Context, string) error
	CreateSession(context.Context, entity.UserSession) error
}

type pushServiceInterface interface {
	CreatePushToken(context.Context, request.CreatePushTokenRequest) (entity.PushToken, error)
	UpdatePushToken(context.Context, request.UpdatePushTokenRequest) error
	DeletePushToken(context.Context, string) error
}

type userApp struct {
	app.Transactor
	repository struct {
		user            repository.UserRepository
		payment         repository.UserPaymentRepository
		smsVerification repository.SmsVerificationRepository // TODO(taekyeom) SMS 관련 로직은 별도 app으로 나중에 빼야 할듯?
		taxiCallRequest repository.TaxiCallRepository
	}

	service struct {
		smsSender service.SmsSenderService
		session   sessionServiceInterface
		payment   service.CardPaymentService
		route     service.MapRouteService
		location  service.LocationService
		push      pushServiceInterface
	}

	actor struct {
		taxiCallRequest *taxicall.TaxiCallActorService
	}
}

func (u userApp) SmsVerificationRequest(ctx context.Context, req request.SmsVerificationRequest) (entity.SmsVerification, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	smsVerification := entity.NewSmsVerification(req.StateKey, requestTime, req.Phone)

	err := u.Run(ctx, func(ctx context.Context, db bun.IDB) error {
		if err := u.repository.smsVerification.Create(ctx, db, smsVerification); err != nil {
			return fmt.Errorf("app.User.SmsVerificationRequest: error while create sms verification:\n%w", err)
		}

		if err := u.service.smsSender.SendSms(ctx, req.Phone, smsVerification.VerificationCode); err != nil {
			return fmt.Errorf("app.User.SmsVerificationRequest: error while send sms message:\n%w", err)
		}
		return nil
	})

	if err != nil {
		return entity.SmsVerification{}, err
	}

	return smsVerification, nil
}

func (u userApp) SmsSignin(ctx context.Context, req request.SmsSigninRequest) (entity.User, string, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	var smsVerification entity.SmsVerification
	var err error
	var user entity.User
	var userSession entity.UserSession

	err = u.RunWithNonRollbackError(ctx, value.ErrUserNotFound, func(ctx context.Context, i bun.IDB) error {
		smsVerification, err = u.repository.smsVerification.FindById(ctx, i, req.StateKey)
		if err != nil {
			return fmt.Errorf("app.User.SmsSignin: error while find sms verification:\n %w", err)
		}

		if smsVerification.VerificationCode != req.VerificationCode {
			return fmt.Errorf("app.User.SmsSignin: invalid verification code:\n%w", value.ErrInvalidOperation)
		}
		user, err = u.repository.user.FindByUserUniqueKey(ctx, i, smsVerification.Phone)
		if errors.Is(value.ErrUserNotFound, err) {
			smsVerification.Verified = true
			if err := u.repository.smsVerification.Update(ctx, i, smsVerification); err != nil {
				return fmt.Errorf("app.User.SmsSingin: error while update sms verification:\n%w", err)
			}
			return err
		}
		if err != nil {
			return fmt.Errorf("app.User.SmsSignin: error while find user by unique key\n%w", err)
		}

		// create new session
		userSession = entity.UserSession{
			Id:         utils.MustNewUUID(),
			UserId:     user.Id,
			ExpireTime: requestTime.AddDate(0, 1, 0), // TODO(taekyeom) Configurable expire time
		}

		if err = u.repository.smsVerification.Delete(ctx, i, smsVerification); err != nil {
			return fmt.Errorf("app.User.SmsSignin: error while delete sms verification:\n%w", err)
		}

		// revoke session
		if err = u.service.session.RevokeSessionByUserId(ctx, user.Id); err != nil {
			return fmt.Errorf("app.User.SmsSignin: error while revoke previous session:\n %w", err)
		}

		if err = u.service.session.CreateSession(ctx, userSession); err != nil {
			return fmt.Errorf("app.User.SmsSignin: error while create new session:\n %w", err)
		}

		return nil
	})

	if err != nil {
		return entity.User{}, "", err
	}

	return user, userSession.Id, nil
}

func (u userApp) Signup(ctx context.Context, req request.UserSignupRequest) (entity.User, string, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	var err error
	var newUser entity.User
	var userSession entity.UserSession

	err = u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		user, err := u.repository.user.FindByUserUniqueKey(ctx, i, req.Phone)
		if !errors.Is(value.ErrUserNotFound, err) && err != nil {
			return fmt.Errorf("app.User.Signup: error while find user by unique key:\n %w", err)
		}

		if user.Id != "" {
			return fmt.Errorf("app.User.Signup: user already exists: %w", value.ErrAlreadyExists)
		}

		smsVerification, err := u.repository.smsVerification.FindById(ctx, i, req.SmsVerificationStateKey)
		if err != nil {
			return fmt.Errorf("app.User.Signup: failed to find sms verification:\n%w", err)
		}

		// check sms verification
		if !smsVerification.Verified {
			return fmt.Errorf("app.User.Signup: not verified phone:\n%w", value.ErrUnAuthorized)
		}

		// create user
		newUser = entity.User{
			Id:            utils.MustNewUUID(),
			FirstName:     req.FirstName,
			LastName:      req.LastName,
			BirthDay:      req.Birthday,
			Phone:         req.Phone,
			Gender:        req.Gender,
			AppOs:         enum.OsTypeFromString(req.AppOs),
			AppVersion:    req.AppVersion,
			AppFcmToken:   req.AppFcmToken,
			UserUniqueKey: req.Phone,
			CreateTime:    requestTime,
			UpdateTime:    requestTime,
			DeleteTime:    time.Time{},
		}
		if err = u.repository.user.CreateUser(ctx, i, newUser); err != nil {
			return fmt.Errorf("app.User.Signup: error while create user:\n %w", err)
		}

		// Create push token
		if _, err := u.service.push.CreatePushToken(ctx, request.CreatePushTokenRequest{
			PrincipalId: newUser.Id,
			FcmToken:    req.AppFcmToken,
		}); err != nil {
			return fmt.Errorf("app.User.Signup: error while create push token: %w", err)
		}

		userSession = entity.UserSession{
			Id:         utils.MustNewUUID(),
			UserId:     newUser.Id,
			ExpireTime: requestTime.AddDate(0, 1, 0), // TODO(taekyeom) Configurable expire time
		}
		if err = u.service.session.CreateSession(ctx, userSession); err != nil {
			return fmt.Errorf("app.User.Signup: error while create new session:\n %w", err)
		}

		// Delete verified sms verification
		if err = u.repository.smsVerification.Delete(ctx, i, smsVerification); err != nil {
			return fmt.Errorf("app.User.Signup: error while delete sms verification:\n %w", err)
		}

		return nil
	})

	if err != nil {
		return entity.User{}, "", err
	}

	return newUser, userSession.Id, nil
}

func (u userApp) UpdateUser(ctx context.Context, req request.UserUpdateRequest) (entity.User, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	var user entity.User
	var err error

	err = u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		user, err = u.repository.user.FindById(ctx, i, req.Id)
		if err != nil {
			return fmt.Errorf("app.user.UpdateUser: error while find user by id:\n %w", err)
		}
		user.AppOs = enum.OsTypeFromString(req.AppOs)
		user.AppVersion = req.AppVersion
		user.AppFcmToken = req.AppFcmToken
		user.UpdateTime = requestTime

		if err = u.repository.user.UpdateUser(ctx, i, user); err != nil {
			return fmt.Errorf("app.user.UpdateUser: error while update user:\n %w", err)
		}

		// Update push token
		if err := u.service.push.UpdatePushToken(ctx, request.UpdatePushTokenRequest{
			PrincipalId: user.Id,
			FcmToken:    req.AppFcmToken,
		}); err != nil {
			return fmt.Errorf("app.User.UpdateUser: error while update push token: %w", err)
		}

		return nil
	})

	if err != nil {
		return entity.User{}, err
	}

	return user, nil
}

func (u userApp) GetUser(ctx context.Context, userId string) (entity.User, error) {

	var user entity.User
	var err error

	err = u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		user, err = u.repository.user.FindById(ctx, i, userId)
		if err != nil {
			return fmt.Errorf("app.user.GetUser: error while find user by id:\n %w", err)
		}

		return nil
	})

	if err != nil {
		return entity.User{}, err
	}

	return user, nil
}

func (u userApp) DeleteUser(ctx context.Context, userId string) error {
	return u.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		user, err := u.repository.user.FindById(ctx, i, userId)
		if err != nil {
			return fmt.Errorf("app.user.DeleteUser: error while find user by id:\n %w", err)
		}

		if err = u.repository.user.DeleteUser(ctx, i, user); err != nil {
			return fmt.Errorf("app.user.DeleteUser: error while delete user:\n %w", err)
		}

		// Delete push token
		if err := u.service.push.DeletePushToken(ctx, userId); err != nil {
			return fmt.Errorf("app.User.DeleteUser: error while delete push token: %w", err)
		}

		return nil
	})
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
