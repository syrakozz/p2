package openai

import (
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"

	"disruptive/lib/common"
	"disruptive/lib/openai"
	"disruptive/lib/users"
	e "disruptive/rest/errors"
)

func getModels(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.openai.getModels")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	claims := users.VerifyClaims(ctx, logCtx, c.Get("user").(*jwt.Token), []string{"ai.user"})
	if !claims.IsVerified {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	models, err := openai.GetModels(ctx, logCtx)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to process query")
	}

	logCtx.Info("get models", fid)
	return c.JSON(http.StatusOK, models)
}
