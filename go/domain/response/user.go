package response

import (
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value/enum"
	"github.com/taco-labs/taco/go/utils/slices"
)

type UserResponse struct {
	Id               string      `json:"id"`
	FirstName        string      `json:"firstName"`
	LastName         string      `json:"lastName"`
	Email            string      `json:"email"`
	BirthDay         string      `json:"birthday"`
	Phone            string      `json:"phone"`
	Gender           string      `json:"gender"`
	AppOs            enum.OsType `json:"appOs"`
	AppVersion       string      `json:"osVersion"`
	DefaultPaymentId string      `json:"defaultPaymentId"`
}

func UserToResponse(user entity.User) UserResponse {
	return UserResponse{
		Id:         user.Id,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Email:      user.Email,
		BirthDay:   user.BirthDay,
		Phone:      user.Phone,
		Gender:     user.Gender,
		AppOs:      user.AppOs,
		AppVersion: user.AppVersion,
	}
}

type UserSignupResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type UserPaymentResponse struct {
	Id                  string `json:"id"`
	UserId              string `json:"userId"`
	Name                string `json:"name"`
	CardCompany         string `json:"cardCompany"`
	RedactedCardNumber  string `json:"redactedCardNumber"`
	CardExpirationYear  string `json:"cardExpirationYear"`
	CardExpirationMonth string `json:"cardExpirationMonth"`
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
	}
}

func UserPaymentsToResponse(userPayments []entity.UserPayment) []UserPaymentResponse {
	return slices.Map(userPayments, UserPaymentToResponse)
}

type ListCardPaymentResponse struct {
	DefaultPaymentId string                `json:"defaultPaymentId"`
	Payments         []UserPaymentResponse `json:"payments"`
}

type DeleteCardPaymentResponse struct {
	PaymentId string `json:"paymentId"`
}
