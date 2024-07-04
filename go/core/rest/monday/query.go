package monday

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"

	"disruptive/lib/common"
	"disruptive/lib/monday"
	"disruptive/lib/users"
	e "disruptive/rest/errors"
)

func postQuery(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.monday.postQuery")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	if claims := users.VerifyServiceAdminClaims(ctx, logCtx, c.Get("user").(*jwt.Token)); !claims.IsVerified {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	body := c.Request().Body
	defer body.Close()

	query, err := io.ReadAll(body)
	if err != nil {
		return e.ErrBad(logCtx, fid, "invalid query")
	}

	res, err := monday.PostQuery(ctx, logCtx, query)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to process query")
	}

	logCtx.Info("monday query", fid)
	return c.JSONBlob(http.StatusOK, res)
}
