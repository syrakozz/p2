package highlevel

import (
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"

	"disruptive/lib/common"
	"disruptive/lib/highlevel"
	"disruptive/lib/users"
	e "disruptive/rest/errors"
)

func getContacts(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.highlevel.getContacts")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	if claims := users.VerifyClaims(ctx, logCtx, c.Get("user").(*jwt.Token), []string{"disruptive.service", "highlevel.user"}); !claims.IsVerified {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	location := c.QueryParam("location")
	query := c.QueryParam("query")

	contacts, err := highlevel.GetContacts(ctx, logCtx, location, query)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get contacts")
	}

	logCtx.Info("get contacts", fid)
	return c.JSON(http.StatusOK, contacts)
}
