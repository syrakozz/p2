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

func getAbout(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.rocketreach.getAbout")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	if claims := users.VerifyClaims(ctx, logCtx, c.Get("user").(*jwt.Token), []string{"disruptive.service"}); !claims.IsVerified {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	about, err := rocketreach.GetAbout(ctx, logCtx)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get about")
	}

	logCtx.Info("get about", fid)
	return c.JSON(http.StatusOK, about)
}
