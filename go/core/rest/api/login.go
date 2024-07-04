package api

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"disruptive/config"
	"disruptive/lib/common"
	"disruptive/lib/users"
	e "disruptive/rest/errors"
)

func postLogin(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.postLogin")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	req := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	logCtx = logCtx.With("username", req.Username)

	u, err := users.GetAndVerifyPassword(ctx, logCtx, req.Username, req.Password)
	if err != nil {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	loginSession := uuid.New().String()
	if err := users.SetLoginSession(ctx, logCtx, req.Username, loginSession); err != nil {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unable to set login session")
	}

	exp := time.Now().Add(12 * time.Hour)
	for _, p := range u.Permissions {
		if strings.Contains(p, "admin") {
			exp = time.Now().Add(2 * time.Hour)
			break
		}
	}

	claims := &users.JWTClaims{
		Username:     req.Username,
		Name:         u.Name,
		Permissions:  u.Permissions,
		LoginSession: loginSession,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(config.VARS.JWTSessionSecret))
	if err != nil {
		return e.Err(logCtx, err, fid, "Unable to create login token")
	}

	logCtx.Info("login", fid)

	c.SetCookie(&http.Cookie{Name: "authorization", Value: t, HttpOnly: true, Domain: config.VARS.Domain})
	return c.JSON(http.StatusOK, map[string]string{"token": t})
}

func postLogout(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.postLogout")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	req := struct {
		Username string `json:"username"`
	}{}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	if err := users.SetLoginSession(ctx, logCtx, req.Username, ""); err != nil {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unable to set login session")
	}

	logCtx.Info("logout", fid, "username", req.Username)
	return c.NoContent(http.StatusOK)
}
