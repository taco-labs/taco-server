package service

import (
	"context"
	"errors"

	"github.com/taco-labs/taco/go/domain/request"
)

type accessTokenRequest struct {
	RestApiKey    string `json:"imp_key"`
	RestApiSecret string `json:"imp_secret"`
}

type accessToken struct {
	AccessToken string `json:"access_token"`
}

type UserIdentity struct {
	Gender        string
	BirthDay      string
	Phone         string
	UserUniqueKey string
}

type UserIdentityService interface {
	GetUserIdentity(context.Context, string) (UserIdentity, error)
}

// type iamportService struct {
// 	apiKey accessTokenRequest
// 	cpid   string
// 	client *resty.Client

// 	accessToken *accessToken
// }

// func (i *iamportService) GetUserIdentity(ctx context.Context, validationKey string) (UserIdentity, error) {
// 	i.requestAccessToken(ctx)
// 	return UserIdentity{}, nil
// }

// func (i *iamportService) requestAccessToken(ctx context.Context) error {
// 	resp, err := i.client.R().
// 		SetBody(i.apiKey).
// 		Post("users/getToken")

// 	if err != nil {
// 		return err
// 	}
// 	println(resp)

// 	return nil
// }

// func NewIamportService(endpoint, restApiKey, restApiSecret, cpid string) iamportService {
// 	client := resty.New().SetBaseURL(endpoint)
// 	return iamportService{
// 		apiKey: accessTokenRequest{
// 			RestApiKey:    restApiKey,
// 			RestApiSecret: restApiSecret,
// 		},
// 		cpid:   cpid,
// 		client: client,
// 	}
// }

type mockGenderKey struct{}
type mockBirthdayKey struct{}
type mockPhoneKey struct{}

type mockUserIdentityService struct{}

func (i mockUserIdentityService) GetUserIdentity(ctx context.Context, identityKey string) (UserIdentity, error) {
	mockGender, ok := ctx.Value(mockGenderKey{}).(string)
	if !ok {
		return UserIdentity{}, errors.New("mock validation service needs MockGender body")
	}

	mockBirthday, ok := ctx.Value(mockBirthdayKey{}).(string)
	if !ok {
		return UserIdentity{}, errors.New("mock validation service needs MockBirthday body")
	}

	mockPhone, ok := ctx.Value(mockPhoneKey{}).(string)
	if !ok {
		return UserIdentity{}, errors.New("mock validation service needs MockPhone body")
	}

	return UserIdentity{
		Gender:        mockGender,
		BirthDay:      mockBirthday,
		Phone:         mockPhone,
		UserUniqueKey: identityKey,
	}, nil
}

func NewMockIdentityService() mockUserIdentityService {
	return mockUserIdentityService{}
}

func SetMockIdentity(ctx context.Context, mockIdentity request.MockUserIdentity) context.Context {
	ctx = context.WithValue(ctx, mockGenderKey{}, mockIdentity.MockGender)
	ctx = context.WithValue(ctx, mockBirthdayKey{}, mockIdentity.MockBirthday)
	ctx = context.WithValue(ctx, mockPhoneKey{}, mockIdentity.MockPhone)
	return ctx
}
