package mailfinder

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"disruptive/lib/mailfinder"
	e "disruptive/rest/errors"
)

func getAccount(c echo.Context) error {
	fid := slog.String("fid", "rest.mailfinder.getAccount")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	account, err := mailfinder.GetAccount(c.Request().Context(), logCtx)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to retrieve account")
	}

	logCtx.Info("mailfinder account", fid, "left", account.CreditsLeft)
	return c.JSON(http.StatusOK, account)
}
