package response

import (
	"time"

	"github.com/taco-labs/taco/go/domain/entity"
)

type SmsVerificationRequestResponse struct {
	Id         string    `json:"id"`
	ExpireTime time.Time `json:"expireTime"`
}

type UserResponse struct {
	Id         string `json:"id"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	BirthDay   string `json:"birthday"`
	Phone      string `json:"phone"`
	Gender     string `json:"gender"`
	AppOs      string `json:"appOs"`
	AppVersion string `json:"osVersion"`
}

func UserToResponse(user entity.User) UserResponse {
	return UserResponse{
		Id:         user.Id,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		BirthDay:   user.BirthDay,
		Phone:      user.Phone,
		Gender:     user.Gender,
		AppOs:      string(user.AppOs),
		AppVersion: user.AppVersion,
	}
}

type UserSignupResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type UserPaymentResponse struct {
	Id                  string    `json:"id"`
	UserId              string    `json:"userId"`
	Name                string    `json:"name"`
	CardCompany         string    `json:"cardCompany"`
	RedactedCardNumber  string    `json:"redactedCardNumber"`
	CardExpirationYear  string    `json:"cardExpirationYear"`
	CardExpirationMonth string    `json:"cardExpirationMonth"`
	DefaultPayment      bool      `json:"defaultPayment"`
	CreateTime          time.Time `json:"createTime"`
}

func UserPaymentToResponse(userPayment entity.UserPayment) UserPaymentResponse {
	return UserPaymentResponse{
		Id:                  userPayment.Id,
		UserId:              userPayment.UserId,
		Name:                userPayment.Name,
		CardCompany:         userPayment.CardCompany,
		RedactedCardNumber:  userPayment.RedactedCardNumber,
		CardExpirationYear:  userPayment.CardExpirationYear,
		CardExpirationMonth: userPayment.CardExpirationMonth,
		DefaultPayment:      userPayment.DefaultPayment,
		CreateTime:          userPayment.CreateTime,
	}
}

type ListCardPaymentResponse struct {
	DefaultPaymentId string                `json:"defaultPaymentId"`
	Payments         []UserPaymentResponse `json:"payments"`
}

type DeleteCardPaymentResponse struct {
	PaymentId string `json:"paymentId"`
}
