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
	"github.com/taco-labs/taco/go/domain/value/analytics"
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

type driverAppInterface interface {
	GetDriverByUserUniqueKey(ctx context.Context, uniqueKey string) (entity.DriverDto, error)
}

type pushServiceInterface interface {
	CreatePushToken(context.Context, request.CreatePushTokenRequest) (entity.PushToken, error)
	UpdatePushToken(context.Context, request.UpdatePushTokenRequest) error
	DeletePushToken(context.Context, string) error
}

type taxiCallInterface interface {
	ListTags(context.Context) ([]value.Tag, error)
	ListUserTaxiCallRequest(context.Context, request.ListUserTaxiCallRequest) ([]entity.TaxiCallRequest, string, error)
	LatestUserTaxiCallRequest(context.Context, string) (entity.UserLatestTaxiCallRequest, error)
	CreateTaxiCallRequest(context.Context, string, request.CreateTaxiCallRequest) (entity.TaxiCallRequest, error)
	UserCancelTaxiCallRequest(context.Context, string, request.CancelTaxiCallRequest) error
}

type userPaymentInterface interface {
	CreateUserPayment(ctx context.Context, userPayment entity.UserPayment) error
	GetCardRegistrationRequestParam(context.Context, entity.User) (value.PaymentRegistrationRequestParam, error)
	RegistrationCallback(context.Context, request.PaymentRegistrationCallbackRequest) (entity.UserPayment, error)
	ListUserPayment(context.Context, string) ([]entity.UserPayment, error)
	TryRecoverUserPayment(context.Context, string, string) error
	DeleteUserPayment(context.Context, entity.User, string) error
	BatchDeleteUserPayment(context.Context, entity.User) error
	PaymentTransactionSuccessCallback(ctx context.Context, req request.PaymentTransactionSuccessCallbackRequest) error
	PaymentTransactionFailCallback(ctx context.Context, req request.PaymentTransactionFailCallbackRequest) error
	CreateUserPaymentPoint(ctx context.Context, userPaymentPoint entity.UserPaymentPoint) error
	GetUserPaymentPoint(ctx context.Context, userId string) (entity.UserPaymentPoint, error)
	CreateUserReferral(ctx context.Context, fromUserId string, toUserId string) error
	AddDriverReferralReward(ctx context.Context, driverId string) error
}

type userApp struct {
	app.Transactor
	repository struct {
		user            repository.UserRepository
		smsVerification repository.SmsVerificationRepository // TODO(taekyeom) SMS 관련 로직은 별도 app으로 나중에 빼야 할듯?
		analytics       repository.AnalyticsRepository
	}

	service struct {
		driver      driverAppInterface
		smsSender   service.SmsSenderService
		session     sessionServiceInterface
		mapService  service.MapService
		push        pushServiceInterface
		taxiCall    taxiCallInterface
		userPayment userPaymentInterface
	}
}

func (u userApp) SmsVerificationRequest(ctx context.Context, req request.SmsVerificationRequest) (entity.SmsVerification, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	var smsVerification entity.SmsVerification
	if req.Phone == entity.MockAccountPhone {
		smsVerification = entity.NewMockSmsVerification(req.StateKey, requestTime)
	} else {
		smsVerification = entity.NewSmsVerification(req.StateKey, requestTime, req.Phone)
	}

	err := u.Run(ctx, func(ctx context.Context, db bun.IDB) error {
		if err := u.repository.smsVerification.Create(ctx, db, smsVerification); err != nil {
			return fmt.Errorf("app.User.SmsVerificationRequest: error while create sms verification:\n%w", err)
		}

		if err := u.service.smsSender.SendSms(ctx, req.Phone, smsVerification.VerficationMessage()); err != nil {
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
	referralType := enum.ReferralType_Unknown
	referralId := ""

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
			return fmt.Errorf("app.User.Signup: not verified phone:\n%w", value.ErrInvalidOperation)
		}

		// TODO (taekyeom) mock account seperation
		var userId string
		if smsVerification.MockAccountPhone() {
			userId = value.MockUserId
		} else {
			userId = utils.MustNewUUID()
		}

		// create user
		newUser = entity.User{
			Id:            userId,
			FirstName:     req.FirstName,
			LastName:      req.LastName,
			BirthDay:      req.Birthday,
			Phone:         req.Phone,
			Gender:        req.Gender,
			AppOs:         enum.OsTypeFromString(req.AppOs),
			AppVersion:    req.AppVersion,
			UserUniqueKey: req.Phone,
			CreateTime:    requestTime,
			UpdateTime:    requestTime,
			DeleteTime:    time.Time{},
		}
		if err = u.repository.user.CreateUser(ctx, i, newUser); err != nil {
			return fmt.Errorf("app.User.Signup: error while create user:\n %w", err)
		}

		if req.ReferralCode != "" && !newUser.MockAccount() {
			referralCode, err := value.DecodeReferralCode(req.ReferralCode)
			if err != nil {
				return fmt.Errorf("app.User.Signup: invalid refferral code: %w (%v)", value.ErrInvalidReferralCode, err)
			}

			switch referralCode.ReferralType {
			case enum.ReferralType_User:
				referralUser, err := u.repository.user.FindByUserUniqueKey(ctx, i, referralCode.PhoneNumber)
				if errors.Is(err, value.ErrNotFound) {
					return fmt.Errorf("app.User.Signup: user not found: %w",
						value.NewTacoError(value.ERR_NOTFOUND_REFERRAL_CODE, "user referral not found"))
				}
				if err != nil {
					return fmt.Errorf("app.User.Signup: error while get referral user: %w", err)
				}

				if !referralUser.MockAccount() {
					if err := u.service.userPayment.CreateUserReferral(ctx, newUser.Id, referralUser.Id); err != nil {
						return fmt.Errorf("app.User.Signup: error while create user referral: %w", err)
					}
					referralType = enum.ReferralType_User
					referralId = referralUser.Id
				}
			case enum.ReferralType_Driver:
				referralDriver, err := u.service.driver.GetDriverByUserUniqueKey(ctx, referralCode.PhoneNumber)
				if errors.Is(err, value.ErrNotFound) {
					return fmt.Errorf("app.User.Signup: driver not found: %w",
						value.NewTacoError(value.ERR_NOTFOUND_REFERRAL_CODE, "driver referral not found"))
				}
				if err != nil {
					return fmt.Errorf("app.User.Signup: error while get driver by unique key: %w", err)
				}

				if !referralDriver.MockAccount() {
					if err := u.service.userPayment.AddDriverReferralReward(ctx, referralDriver.Id); err != nil {
						return fmt.Errorf("app.User.Signup: error while update driver referral: %w", err)
					}
					referralType = enum.ReferralType_Driver
					referralId = referralDriver.Id
				}
			}
		}

		// If user account is mock, create mock payment
		if newUser.MockAccount() {
			mockPayment := entity.NewMockPayment(newUser.Id, requestTime)
			if err := u.service.userPayment.CreateUserPayment(ctx, mockPayment); err != nil {
				return fmt.Errorf("app.User.Signup: error while create mock payment: %w", err)
			}
		}

		// Create push token
		if _, err := u.service.push.CreatePushToken(ctx, request.CreatePushTokenRequest{
			PrincipalId: newUser.Id,
			FcmToken:    req.AppFcmToken,
		}); err != nil {
			return fmt.Errorf("app.User.Signup: error while create push token: %w", err)
		}

		// Create user payment point
		if err := u.service.userPayment.CreateUserPaymentPoint(ctx, entity.UserPaymentPoint{
			UserId: newUser.Id,
			Point:  0,
		}); err != nil {
			return fmt.Errorf("app.User.Signup: error while create user payment point: %w", err)
		}

		userSession = entity.UserSession{
			Id:         utils.MustNewUUID(),
			UserId:     newUser.Id,
			ExpireTime: requestTime.AddDate(0, 1, 0),
		}
		if err = u.service.session.CreateSession(ctx, userSession); err != nil {
			return fmt.Errorf("app.User.Signup: error while create new session:\n %w", err)
		}

		// Delete verified sms verification
		if err = u.repository.smsVerification.Delete(ctx, i, smsVerification); err != nil {
			return fmt.Errorf("app.User.Signup: error while delete sms verification:\n %w", err)
		}

		userSignupAnalytics := entity.NewAnalytics(requestTime, analytics.UserSignupPayload{
			UserId:       newUser.Id,
			ReferralType: referralType,
			ReferralId:   referralId,
		})
		if err := u.repository.analytics.Create(ctx, i, userSignupAnalytics); err != nil {
			return fmt.Errorf("app.User.Signup: error while create analytics event: %w", err)
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

		userPaymentPoint, err := u.service.userPayment.GetUserPaymentPoint(ctx, userId)
		if err != nil {
			return fmt.Errorf("app.user.GetUser: error while find user payment point:\n %w", err)
		}

		user.UserPoint = userPaymentPoint.Point

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

		if err := u.service.userPayment.BatchDeleteUserPayment(ctx, user); err != nil {
			return fmt.Errorf("app.user.DeleteUser: error while batch delete user payment:\n %w", err)
		}

		// Delete push token
		if err := u.service.push.DeletePushToken(ctx, userId); err != nil {
			return fmt.Errorf("app.User.DeleteUser: error while delete push token: %w", err)
		}

		return nil
	})
}

func NewUserApp(opts ...userAppOption) (*userApp, error) {
	ua := &userApp{}

	for _, opt := range opts {
		opt(ua)
	}

	return ua, ua.validateApp()
}
