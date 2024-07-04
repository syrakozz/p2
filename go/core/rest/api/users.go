package api

import (
	"log/slog"
	"net/http"
	"net/mail"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"

	"disruptive/lib/common"
	"disruptive/lib/users"
	e "disruptive/rest/errors"
)

func postUsers(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.postUsers")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	claims := users.VerifyServiceAdminClaims(ctx, logCtx, c.Get("user").(*jwt.Token))
	if !claims.IsVerified {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	req := users.User{}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read user data")
	}

	logCtx = logCtx.With("username", req.Username)

	if _, err := mail.ParseAddress(req.Username); err != nil {
		return e.ErrBad(logCtx, fid, "invalid username")
	}

	permsMap := map[string][]string{}
	for _, p := range req.Permissions {
		parts := strings.Split(p, ".")
		if len(parts) < 1 {
			continue
		}

		// a service by itself gives list permissions
		permsMap[parts[0]] = []string{parts[0]}
		permsMap[p] = parts
	}

	// If claim is only ServiceAdmin,
	// only allow add user with service permissions.
	if !claims.IsAdmin {
		serviceAdmins := map[string]bool{}
		for _, p := range claims.Permissions {
			parts := strings.Split(p, ".")
			if len(parts) == 2 && parts[1] == "admin" {
				serviceAdmins[parts[0]] = true
			}
		}

		for _, parts := range permsMap {
			if !serviceAdmins[parts[0]] {
				return e.ErrBad(logCtx, fid, "invalid permissions")
			}
		}
	}

	permissions := make([]string, 0, len(permsMap))
	for p := range permsMap {
		permissions = append(permissions, p)
	}
	req.Permissions = permissions

	if err := users.Add(ctx, logCtx, req); err != nil {
		return e.Err(logCtx, err, fid, "unable to add user")
	}

	logCtx.Info("add user")
	return c.NoContent(http.StatusCreated)
}

func getUsers(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.getUsers")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	claims := users.VerifyServiceAdminClaims(ctx, logCtx, c.Get("user").(*jwt.Token))
	if !claims.IsVerified {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	service := c.QueryParam("service")
	if service == "" {
		return e.ErrBad(logCtx, fid, "missing service")
	}

	// If claim is only ServiceAdmin,
	// only allow retrieving users with allowed service permissions.
	if !claims.IsAdmin {
		serviceAdmins := map[string]bool{}
		for _, p := range claims.Permissions {
			parts := strings.Split(p, ".")
			if len(parts) == 2 && parts[1] == "admin" {
				serviceAdmins[parts[0]] = true
			}
		}

		if !serviceAdmins[service] {
			return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
		}
	}

	users, err := users.GetUsersByPermission(ctx, logCtx, service)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to query users")
	}

	logCtx.Info("get users", fid)
	return c.JSON(http.StatusOK, users)
}

func getUserMe(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.getUserMe")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	claims, ok := users.GetClaims(c.Get("user").(*jwt.Token))
	if !ok {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	username := claims.Username
	logCtx = logCtx.With("username", username)

	user, err := users.Get(ctx, logCtx, username)
	if err != nil {
		return e.Err(logCtx, err, fid, "Not Found")
	}

	logCtx.Info("get user", fid)
	return c.JSON(http.StatusOK, user)
}

func getUser(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.getUser")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	claims := users.VerifyServiceAdminClaims(ctx, logCtx, c.Get("user").(*jwt.Token))
	if !claims.IsVerified {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	username := c.Param("username")
	logCtx = logCtx.With("username", username)

	user, err := users.Get(ctx, logCtx, username)
	if err != nil {
		return e.Err(logCtx, err, fid, "Not Found")
	}

	logCtx.Info("get user", fid)
	return c.JSON(http.StatusOK, user)
}

func patchUser(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.patchUsers")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	claims := users.VerifyServiceAdminClaims(ctx, logCtx, c.Get("user").(*jwt.Token))
	if !claims.IsVerified {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	username := c.Param("username")
	logCtx = logCtx.With("username", username)

	req := users.User{Username: username}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read user data")
	}

	permissionsMap := map[string][]string{}
	for _, p := range req.Permissions {
		parts := strings.Split(p, ".")
		if len(parts) < 1 || len(parts[0]) < 1 {
			continue
		}

		if parts[0][0] != '-' {
			permissionsMap[parts[0]] = []string{parts[0]}
		}
		permissionsMap[p] = parts
	}

	// If claim is only ServiceAdmin,
	// only allow add user with service permissions.
	if !claims.IsAdmin {
		serviceAdmins := map[string]bool{}
		for _, p := range claims.Permissions {
			parts := strings.Split(p, ".")
			if len(parts) == 2 && parts[1] == "admin" {
				serviceAdmins[parts[0]] = true
			}
		}

		for _, parts := range permissionsMap {
			if !serviceAdmins[parts[0]] && !(parts[0][0] == '-' && serviceAdmins[parts[0][1:]]) {
				return e.ErrBad(logCtx, fid, "invalid permissions")
			}
		}
	}

	permissions := make([]string, 0, len(permissionsMap))
	for p, parts := range permissionsMap {
		if len(parts) == 1 && parts[0][0] == '-' {
			continue
		}
		permissions = append(permissions, p)
	}
	req.Permissions = permissions

	if err := users.Modify(ctx, logCtx, req); err != nil {
		return e.Err(logCtx, err, fid, "unable to modify user")
	}

	logCtx.Info("modify user", fid)
	return c.NoContent(http.StatusOK)

}

func deleteUser(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.deleteUser")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	claims := users.VerifyServiceAdminClaims(ctx, logCtx, c.Get("user").(*jwt.Token))
	if !claims.IsVerified {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	username := c.Param("username")
	logCtx = logCtx.With("username", username)

	user, err := users.Get(ctx, logCtx, username)
	if err != nil {
		return e.Err(logCtx, err, fid, "Not Found")
	}

	// If claim is only ServiceAdmin,
	// only allow add user with service permissions.
	if !claims.IsAdmin {
		serviceAdmins := map[string]bool{}
		for _, p := range claims.Permissions {
			parts := strings.Split(p, ".")
			if len(parts) == 2 && parts[1] == "admin" {
				serviceAdmins[parts[0]] = true
			}
		}

		for _, p := range user.Permissions {
			parts := strings.Split(p, ".")
			if len(parts) < 1 {
				continue
			}
			if !serviceAdmins[parts[0]] {
				return e.ErrBad(logCtx, fid, "invalid permissions")
			}
		}
	}

	if err := users.Delete(ctx, logCtx, username); err != nil {
		return e.Err(logCtx, err, fid, "unable to modify user")
	}

	logCtx.Info("delete user", fid)
	return c.NoContent(http.StatusOK)
}
