package server

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/utils"
	"go.uber.org/zap"
)

// TODO (taekyeom) Do better error handler
func ToResponse(e echo.Context, err error) error {
	if err == nil {
		return err
	}

	ctx := e.Request().Context()
	logger := utils.GetLogger(ctx)

	tacoError := &value.TacoError{}
	if !errors.As(err, tacoError) {
		logger.Error("Non taco error occurred",
			zap.Error(err),
			zap.String("path", e.Request().RequestURI),
		)
		return err
	}

	herr := echo.NewHTTPError(http.StatusInternalServerError)
	logger.Error("Taco error occurred",
		zap.Error(err),
		zap.String("code", string(tacoError.ErrCode)),
		zap.String("message", tacoError.Message),
		zap.String("path", e.Path()),
		zap.String("uri", e.Request().RequestURI),
		zap.String("method", e.Request().Method),
		zap.Any("query", e.QueryParams()),
	)

	switch tacoError.ErrCode {
	case value.ERR_UNAUTHENTICATED, value.ERR_UNAUTHORIZED, value.ERR_SESSION_EXPIRED:
		err := echo.NewHTTPError(http.StatusBadRequest, tacoError)
		herr.SetInternal(err)
	case value.ERR_DB_INTERNAL, value.ERR_INTERNAL:
		err := echo.NewHTTPError(http.StatusInternalServerError, tacoError)
		herr.SetInternal(err)
	case value.ERR_NOTFOUND:
		err := echo.NewHTTPError(http.StatusNotFound, tacoError)
		herr.SetInternal(err)
	case value.ERR_EXTERNAL, value.ERR_EXTERNAL_PAYMENT:
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
		err := echo.NewHTTPError(http.StatusInternalServerError, tacoError)
		herr.SetInternal(err)
	}

	return herr
}
