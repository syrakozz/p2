package mailfinder

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"disruptive/lib/mailfinder"
	e "disruptive/rest/errors"
)

func getPing(c echo.Context) error {
	fid := slog.String("fid", "rest.mailfinder.getPing")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	pong, err := mailfinder.GetStatus(c.Request().Context(), logCtx)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to ping mailfinder")
	}

	logCtx.Info("mailfinder ping", fid, "pong", pong)
	return c.JSON(http.StatusOK, map[string]bool{"pong": pong})
}
