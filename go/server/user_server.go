package server

import (
	"context"
	"net/http"

	"github.com/ktk1012/taco/go/domain/entity"
	"github.com/ktk1012/taco/go/domain/request"
	"github.com/ktk1012/taco/go/domain/response"
	"github.com/ktk1012/taco/go/service"
	"github.com/labstack/echo/v4"
)

type UserApp interface {
	Signup(context.Context, request.UserSignupRequest) (entity.User, string, error)
}

func (u userServer) Signup(e echo.Context) error {
	ctx := e.Request().Context()

	// TODO(taekyeom) Remove mock
	type mockReq struct {
		request.UserSignupRequest
		request.MockUserIdentity
	}

	req := mockReq{}

	if err := e.Bind(&req); err != nil {
		// TODO(taekyeom) Error handle
		return e.String(http.StatusBadRequest, "bind error 1")
	}

	ctx = service.SetMockIdentity(ctx, req.MockUserIdentity)

	user, token, err := u.app.user.Signup(ctx, req.UserSignupRequest)

	resp := response.UserSignupResponse{
		Token: token,
		User:  response.UserResponseFromUser(user),
	}

	if err != nil {
		// TODO(taekyeom) Error handle
		return e.String(http.StatusBadRequest, err.Error())
	}

	return e.JSON(http.StatusOK, resp)
}
