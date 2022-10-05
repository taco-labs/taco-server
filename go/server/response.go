package server

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/utils"
)

// TODO (taekyeom) Do better error handler
func ToResponse(err error) error {
	if err == nil {
		return err
	}

	defer utils.Logger.Sync()
	sugar := utils.Logger.Sugar()

	tacoError := &value.TacoError{}
	if !errors.As(err, tacoError) {
		sugar.Errorw("Non taco error occurred",
			"error", err.Error(),
		)
		return err
	}

	herr := echo.NewHTTPError(http.StatusInternalServerError)
	sugar.Errorw("Taco error occurred",
		"error", err.Error(),
		"code", tacoError.ErrCode,
		"message", tacoError.Message,
	)

	switch tacoError.ErrCode {
	case value.ERR_UNAUTHENTICATED, value.ERR_UNAUTHORIZED, value.ERR_SESSION_EXPIRED:
		err := echo.NewHTTPError(http.StatusBadRequest)
		herr.SetInternal(err)
	case value.ERR_DB_INTERNAL, value.ERR_INTERNAL:
		err := echo.NewHTTPError(http.StatusInternalServerError, tacoError)
		herr.SetInternal(err)
	case value.ERR_NOTFOUND:
		err := echo.NewHTTPError(http.StatusNotFound, tacoError)
		herr.SetInternal(err)
	case value.ERR_EXTERNAL:
		err := echo.NewHTTPError(http.StatusInternalServerError, tacoError)
		herr.SetInternal(err)
	case value.ERR_ALREADY_EXISTS:
		err := echo.NewHTTPError(http.StatusForbidden, tacoError)
		herr.SetInternal(err)
	case value.ERR_INVALID:
		err := echo.NewHTTPError(http.StatusForbidden, tacoError)
		herr.SetInternal(err)
	case value.ERR_NEED_CONFIRMATION:
		err := echo.NewHTTPError(http.StatusNotAcceptable, tacoError)
		herr.SetInternal(err)
	case value.ERR_UNSUPPORTED:
		err := echo.NewHTTPError(http.StatusNotAcceptable, tacoError)
		herr.SetInternal(err)
	default:
		err := echo.NewHTTPError(http.StatusInternalServerError, err)
		herr.SetInternal(err)
	}

	return herr
}
