package rocketreach

import (
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"

	"disruptive/lib/common"
	"disruptive/lib/rocketreach"
	"disruptive/lib/users"
	e "disruptive/rest/errors"
)

func postSearch(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.rocketreach.postSearch")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	if claims := users.VerifyClaims(ctx, logCtx, c.Get("user").(*jwt.Token), []string{"disruptive.service", "rocketreach.user"}); !claims.IsVerified {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	req := map[string]any{}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	search, err := rocketreach.PostSearch(ctx, logCtx, req)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to post search")
	}

	logCtx.Info("post search", fid)
	return c.JSON(http.StatusOK, search)
}
