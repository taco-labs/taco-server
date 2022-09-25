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
		user            repository.UserRepository
		payment         repository.UserPaymentRepository
		smsVerification repository.SmsVerificationRepository // TODO(taekyeom) SMS 관련 로직은 별도 app으로 나중에 빼야 할듯?
		taxiCallRequest repository.TaxiCallRepository
	}

	service struct {
		smsSender service.SmsSenderService
		session   SessionInterface
		payment   service.CardPaymentService
	}

	actor struct {
	}
}

func (u userApp) SmsVerificationRequest(ctx context.Context, req request.SmsVerificationRequest) (entity.SmsVerification, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	ctx, err := u.Start(ctx)
	if err != nil {
		return entity.SmsVerification{}, err
	}
	defer func() {
		err = u.Done(ctx, err)
	}()

	smsVerification := entity.NewSmsVerification(req.StateKey, requestTime, req.Phone)

	if err = u.repository.smsVerification.Create(ctx, smsVerification); err != nil {
		return entity.SmsVerification{}, fmt.Errorf("app.User.SmsVerificationRequest: error while create sms verification:\n%w", err)
	}

	if err = u.service.smsSender.SendSms(ctx, req.Phone, smsVerification.VerificationCode); err != nil {
		return entity.SmsVerification{}, fmt.Errorf("app.User.SmsVerificationRequest: error while send sms message:\n%w", err)
	}

	return smsVerification, nil
}

func (u userApp) SmsSignin(ctx context.Context, req request.SmsSigninRequest) (entity.User, string, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	ctx, err := u.Start(ctx)
	if err != nil {
		return entity.User{}, "", err
	}
	defer func() {
		// XXX (taekyeom) 유저가 없는 경우에 대해서 adhoc 하게 핸들링 하고 있음.. 더 나은 방향으로 개선 필요함
		if errors.Is(value.ErrUserNotFound, err) {
			err = u.Done(ctx, nil)
		} else {
			err = u.Done(ctx, err)
		}
	}()

	// First check sms code
	smsVerification, err := u.repository.smsVerification.FindById(ctx, req.StateKey)
	if err != nil {
		return entity.User{}, "", fmt.Errorf("app.User.SmsSignin: error while find sms verification:\n %w", err)
	}

	if smsVerification.ExpireTime.Before(requestTime) {
		return entity.User{}, "", fmt.Errorf("app.User.SmsSignin: expired:\n%w", value.ErrInvalidOperation)
	}

	if smsVerification.VerificationCode != req.VerificationCode {
		return entity.User{}, "", fmt.Errorf("app.User.SmsSignin: invalid verification code:\n%w", value.ErrInvalidOperation)
	}

	user, err := u.repository.user.FindByUserUniqueKey(ctx, smsVerification.Phone)
	if errors.Is(value.ErrUserNotFound, err) {
		smsVerification.Verified = true
		if err = u.repository.smsVerification.Update(ctx, smsVerification); err != nil {
			return entity.User{}, "", fmt.Errorf("app.User.SmsSingin: error while update sms verification:\n%w", err)
		}
		return entity.User{}, "", fmt.Errorf("app.User.SmsSignin: user not found\n%w", value.ErrUserNotFound)
	}
	if err != nil {
		return entity.User{}, "", fmt.Errorf("app.User.SmsSignin: error while find user by unique key:\n%w", err)
	}

	if err = u.repository.smsVerification.Delete(ctx, smsVerification); err != nil {
		return entity.User{}, "", fmt.Errorf("app.User.SmsSignin: error while delete sms verification:\n%w", err)
	}

	// revoke session
	if err = u.service.session.RevokeSessionByUserId(ctx, user.Id); err != nil {
		return entity.User{}, "", fmt.Errorf("app.User.SmsSignin: error while revoke previous session:\n %w", err)
	}

	// create new session
	userSession := entity.UserSession{
		Id:         utils.MustNewUUID(),
		UserId:     user.Id,
		ExpireTime: requestTime.AddDate(0, 1, 0), // TODO(taekyeom) Configurable expire time
	}
	if err = u.service.session.CreateSession(ctx, userSession); err != nil {
		return entity.User{}, "", fmt.Errorf("app.User.SmsSignin: error while create new session:\n %w", err)
	}

	return user, userSession.Id, nil
}

func (u userApp) Signup(ctx context.Context, req request.UserSignupRequest) (entity.User, string, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)

	ctx, err := u.Start(ctx)
	if err != nil {
		return entity.User{}, "", err
	}
	defer func() {
		err = u.Done(ctx, err)
	}()

	user, err := u.repository.user.FindByUserUniqueKey(ctx, req.Phone)
	if !errors.Is(value.ErrUserNotFound, err) && err != nil {
		return entity.User{}, "", fmt.Errorf("app.User.Signup: error while find user by unique key:\n %w", err)
	}

	if user.Id != "" {
		return user, "", fmt.Errorf("app.User.Signup: user already exists: %w", value.ErrAlreadyExists)
	}

	smsVerification, err := u.repository.smsVerification.FindById(ctx, req.SmsVerificationStateKey)
	if err != nil {
		return entity.User{}, "", fmt.Errorf("app.User.Signup: failed to find sms verification:\n%w", err)
	}

	// check sms verification
	if !smsVerification.Verified {
		return entity.User{}, "", fmt.Errorf("app.User.Signup: not verified phone:\n%w", value.ErrUnAuthorized)
	}

	// create user
	newUser := entity.User{
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
	if err = u.repository.user.CreateUser(ctx, newUser); err != nil {
		return entity.User{}, "", fmt.Errorf("app.User.Signup: error while create user:\n %w", err)
	}

	userSession := entity.UserSession{
		Id:         utils.MustNewUUID(),
		UserId:     newUser.Id,
		ExpireTime: requestTime.AddDate(0, 1, 0), // TODO(taekyeom) Configurable expire time
	}
	if err = u.service.session.CreateSession(ctx, userSession); err != nil {
		return entity.User{}, "", fmt.Errorf("app.User.Signup: error while create new session:\n %w", err)
	}

	// Delete verified sms verification
	if err = u.repository.smsVerification.Delete(ctx, smsVerification); err != nil {
		return entity.User{}, "", fmt.Errorf("app.User.Signup: error while delete sms verification:\n %w", err)
	}

	return newUser, userSession.Id, nil
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
		return entity.User{}, fmt.Errorf("app.user.UpdateUser: error while find user by id:\n %w", err)
	}

	user.AppOs = enum.OsTypeFromString(req.AppOs)
	user.AppVersion = req.AppVersion
	user.AppFcmToken = req.AppFcmToken
	user.UpdateTime = requestTime

	if err = u.repository.user.UpdateUser(ctx, user); err != nil {
		return entity.User{}, fmt.Errorf("app.user.UpdateUser: error while update user:\n %w", err)
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
		return entity.User{}, fmt.Errorf("app.user.GetUser: error while find user by id:\n %w", err)
	}

	return user, nil
}

func (u userApp) DeleteUser(ctx context.Context, userId string) error {
	ctx, err := u.Start(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = u.Done(ctx, err)
	}()

	user, err := u.repository.user.FindById(ctx, userId)
	if err != nil {
		return fmt.Errorf("app.user.DeleteUser: error while find user by id:\n %w", err)
	}

	if err = u.repository.user.DeleteUser(ctx, user); err != nil {
		return fmt.Errorf("app.user.DeleteUser: error while delete user:\n %w", err)
	}

	return nil
}

func (u userApp) ListTaxiCallRequest(ctx context.Context, userId string) ([]entity.TaxiCallRequest, error) {
	ctx, err := u.Start(ctx)
	if err != nil {
		return []entity.TaxiCallRequest{}, err
	}
	defer func() {
		err = u.Done(ctx, err)
	}()

	taxiCallRequests, err := u.repository.taxiCallRequest.ListByUserId(ctx, userId)
	if err != nil {
		return []entity.TaxiCallRequest{}, fmt.Errorf("app.user.ListTaxiCallRequest: error while get taxi call requests:\n%w", err)
	}

	return taxiCallRequests, nil
}

func (u userApp) GetLatestTaxiCallRequest(ctx context.Context, userId string) (entity.TaxiCallRequest, error) {
	ctx, err := u.Start(ctx)
	if err != nil {
		return entity.TaxiCallRequest{}, err
	}
	defer func() {
		err = u.Done(ctx, err)
	}()

	latestTaxiCallRequest, err := u.repository.taxiCallRequest.GetLatestByUserId(ctx, userId)
	if err != nil {
		return entity.TaxiCallRequest{}, fmt.Errorf("app.user.GetLatestTaxiCall: error while get latest taxi call:\n%w", err)
	}

	return latestTaxiCallRequest, nil
}

func (u userApp) CreateTaxiCallRequest(ctx context.Context, req request.CreateTaxiCallRequest) (entity.TaxiCallRequest, error) {
	requestTime := utils.GetRequestTimeOrNow(ctx)
	userId := utils.GetUserId(ctx)

	ctx, err := u.Start(ctx)
	if err != nil {
		return entity.TaxiCallRequest{}, err
	}
	defer func() {
		err = u.Done(ctx, err)
	}()

	// check latest call
	latestTaxiCallRequest, err := u.repository.taxiCallRequest.GetLatestByUserId(ctx, userId)
	if err != nil && !errors.Is(err, value.ErrNotFound) {
		return entity.TaxiCallRequest{}, fmt.Errorf("app.user.CreateTaxiCallRequest: error while get latest taxi call:\n%w", err)
	}

	isNotFound := errors.Is(err, value.ErrNotFound)
	isActive := latestTaxiCallRequest.CurrentState.Active()

	if !isNotFound && isActive {
		err = fmt.Errorf("app.user.CreateTaxiCallRequest: already active taxi call request exists:\n%w", value.ErrAlreadyExists)
		return entity.TaxiCallRequest{}, err
	}

	// check payment
	userPayment, err := u.repository.payment.GetUserPayment(ctx, req.PaymentId)
	if err != nil {
		return entity.TaxiCallRequest{}, fmt.Errorf("app.user.CreateTaxiCallRequest: error while get user payment:\n%w", err)
	}

	if userPayment.UserId != userId {
		err = fmt.Errorf("app.User.CreateTaxiCallRequest: unaurhorized payment:%w", value.ErrUnAuthorized)
		return entity.TaxiCallRequest{}, err
	}

	// create taxi call request
	initialState := enum.TaxiCallState_Requested
	taxiCallRequest := entity.TaxiCallRequest{
		Id:        utils.MustNewUUID(),
		UserId:    userId,
		Departure: req.Departure,
		Arrival:   req.Arrival,
		PaymentSummary: value.PaymentSummary{
			PaymentId:  userPayment.Id,
			Company:    userPayment.CardCompany,
			CardNumber: userPayment.RedactedCardNumber,
		},
		RequestBasePrice:          req.RequestBasePrice,
		RequestMinAdditionalPrice: 0,                    // TODO(taekyeom) To be paramterized
		RequestMaxAdditionalPrice: req.RequestBasePrice, // TODO(taekyeom) To be paramterized
		CurrentState:              initialState,
		CallHistory: []value.TaxiCallRequestHistory{
			{
				TaxiCallState: initialState,
				CreateTime:    requestTime,
			},
		},
		CreateTime: requestTime,
		UpdateTime: requestTime,
	}

	if err = u.repository.taxiCallRequest.Create(ctx, taxiCallRequest); err != nil {
		return entity.TaxiCallRequest{}, fmt.Errorf("app.user.CreateTaxiCallRequest: error while create taxi call request:%w", err)
	}

	return taxiCallRequest, nil
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

	if u.repository.smsVerification == nil {
		return errors.New("user app need sms verification repository")
	}

	if u.repository.taxiCallRequest == nil {
		return errors.New("user app need taxi call request repository")
	}

	if u.service.session == nil {
		return errors.New("user app need user session repository")
	}

	if u.service.payment == nil {
		return errors.New("user app need card payment service")
	}

	if u.service.smsSender == nil {
		return errors.New("user app need sms sender service")
	}

	return nil
}
