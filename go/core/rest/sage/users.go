package sage

import (
	"log/slog"
	"net/http"
	"net/mail"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"

	"disruptive/lib/common"
	"disruptive/lib/sage/users"
	u "disruptive/lib/users"
	e "disruptive/rest/errors"
)

func postUsers(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.postUsers")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	claims := u.VerifyServiceAdminClaims(ctx, logCtx, c.Get("user").(*jwt.Token))
	if !claims.IsVerified {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	req := struct {
		Username string `json:"username"`
	}{}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	logCtx = logCtx.With("username", req.Username)

	if _, err := mail.ParseAddress(req.Username); err != nil {
		return e.ErrBad(logCtx, fid, "invalid username")
	}

	if err := users.Add(ctx, logCtx, req.Username); err != nil {
		return e.Err(logCtx, err, fid, "unable to add user")
	}

	logCtx.Info("created user", fid)
	return c.NoContent(http.StatusCreated)
}

func getUser(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.sage.getUser")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	claims := u.VerifyClaims(ctx, logCtx, c.Get("user").(*jwt.Token), sageUser)
	if !claims.IsVerified {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	username := ""

	if claims.IsAdmin || claims.IsServiceAdmin {
		username = c.QueryParam("username")
	}

	if username == "" {
		username = claims.Username
	}

	logCtx = logCtx.With("username", username)

	user, err := users.Get(ctx, logCtx, username)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get users")
	}

	logCtx.Info("get user", fid)
	return c.JSON(http.StatusOK, user)
}

func patchUser(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.sage.patchUser")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	claims := u.VerifyClaims(ctx, logCtx, c.Get("user").(*jwt.Token), sageUser)
	if !claims.IsVerified {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	username := claims.Username
	logCtx = logCtx.With("username", username)

	req := users.Document{}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	if err := users.Patch(ctx, logCtx, username, req); err != nil {
		return e.Err(logCtx, err, fid, "unable to update user")
	}

	logCtx.Info("update user", fid)
	return c.NoContent(http.StatusOK)
}
