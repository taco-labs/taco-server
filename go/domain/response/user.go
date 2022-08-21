package response

import (
	"github.com/ktk1012/taco/go/domain/entity"
	"github.com/ktk1012/taco/go/domain/value/enum"
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

func UserResponseFromUser(user entity.User) UserResponse {
	return UserResponse{
		Id:               user.Id,
		FirstName:        user.FirstName,
		LastName:         user.LastName,
		Email:            user.Email,
		BirthDay:         user.BirthDay,
		Phone:            user.Phone,
		Gender:           user.Gender,
		AppOs:            user.AppOs,
		AppVersion:       user.AppVersion,
		DefaultPaymentId: user.DefaultPaymentId.String,
	}
}

type UserSignupResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}
