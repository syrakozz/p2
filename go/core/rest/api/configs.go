package api

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"disruptive/lib/configs"
	"disruptive/rest/auth"
	e "disruptive/rest/errors"
)

func getConfig(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.api.getConfig")

	name := c.Param("config")
	if name == "" {
		return e.ErrBad(logCtx, fid, "invalid config")
	}
	logCtx = logCtx.With("config", name)

	document, err := configs.Get(ctx, logCtx, name)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get config")
	}

	logCtx.Info("get config", fid)
	return c.JSON(http.StatusOK, document)
}
