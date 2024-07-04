package api

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
)

func getPing(c echo.Context) error {
	fid := slog.String("fid", "rest.getPing")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))
	logCtx.Info("ping", fid)
	return c.JSON(http.StatusOK, map[string]bool{"pong": true})
}
