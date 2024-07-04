// Package errors logs an error and returns an Echo error
package errors

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"disruptive/lib/common"
)

// Err handles HTTP errors by switching on the type of error,
// loging it, and returning the correct HTTP error code.
func Err(logCtx *slog.Logger, err error, fid slog.Attr, msg string) *echo.HTTPError {
	switch {
	case errors.Is(err, common.ErrAlreadyExists{}):
		logCtx.Warn("already exists", fid, "error", err)
		return echo.NewHTTPError(http.StatusConflict, "already exists")

	case errors.Is(err, common.ErrBadGateway{}):
		logCtx.Error("bad gateway", fid, "error", err)
		return echo.NewHTTPError(http.StatusBadGateway, "bad gateway: "+err.Error())

	case errors.Is(err, common.ErrBadRequest{}):
		logCtx.Error("bad request", fid, "error", err)
		return echo.NewHTTPError(http.StatusBadRequest, "bad request: "+err.Error())

	case errors.Is(err, common.ErrConnection{}):
		logCtx.Error("unable to connect", fid, "error", err)
		return echo.NewHTTPError(http.StatusBadGateway, "unable to connect: "+err.Error())

	case errors.Is(err, common.ErrConstraintViolation{}):
		logCtx.Error("constraint violation", fid, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "constraint violation: "+err.Error())

	case errors.Is(err, common.ErrContextValues{}):
		logCtx.Error("context values", fid, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "unable to set context values: "+err.Error())

	case errors.Is(err, common.ErrForbidden{}):
		logCtx.Error("forbidden", fid, "error", err)
		return echo.NewHTTPError(http.StatusForbidden, "forbidden: "+err.Error())

	case errors.Is(err, common.ErrGatewayTimeout{}):
		logCtx.Error("gateway timeout", fid, "error", err)
		return echo.NewHTTPError(http.StatusGatewayTimeout, "gateway timeout: "+err.Error())

	case errors.Is(err, common.ErrGone{}):
		logCtx.Error("gone", fid, "error", err)
		return echo.NewHTTPError(http.StatusGone, "gone: "+err.Error())

	case errors.Is(err, common.ErrIAPUnauthorized):
		logCtx.Error("IAP unauthorized", fid, "error", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "IAP unauthorized")

	case errors.Is(err, common.ErrNotFound{}):
		logCtx.Warn("not found", fid, "error", err)
		return echo.NewHTTPError(http.StatusNotFound, "not found")

	case errors.Is(err, common.ErrNoResults): // TG specific. TG can "successfully" return no results.
		logCtx.Warn("no results", fid, "error", err)
		return echo.NewHTTPError(http.StatusNotFound, "no results")

	case errors.Is(err, common.ErrPaymentRequired{}):
		logCtx.Error("payment required", fid, "error", err)
		return echo.NewHTTPError(http.StatusPaymentRequired, "payment required: "+err.Error())

	case errors.Is(err, common.ErrPreconditionFailed{}):
		logCtx.Error("precondition failed", fid, "error", err)
		return echo.NewHTTPError(http.StatusPreconditionFailed, "precondition failed: "+err.Error())

	case errors.Is(err, common.ErrTooManyRequests{}):
		logCtx.Error("too many requests", fid, "error", err)
		return echo.NewHTTPError(http.StatusTooManyRequests, "too many requests: "+err.Error())

	case errors.Is(err, common.ErrUnauthorized):
		logCtx.Error("unauthorized", fid, "error", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")

	case errors.Is(err, common.ErrUnprocessable{}):
		logCtx.Error("unprocessable", fid, "error", err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "unprocessable")

	default:
		logCtx.Error(msg, fid, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, msg+": "+err.Error())
	}
}

// ErrBad is an HTTP bad request helper function.
// Log the message and return an HTTP 400 error.
func ErrBad(logCtx *slog.Logger, fid slog.Attr, msg string) *echo.HTTPError {
	logCtx.Warn(msg, fid)
	return echo.NewHTTPError(http.StatusBadRequest, msg)
}
