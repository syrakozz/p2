package sage

import (
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"

	"disruptive/lib/common"
	"disruptive/lib/sage/sessions"
	u "disruptive/lib/users"
	e "disruptive/rest/errors"
)

func postSessions(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.sage.postSessions")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	claims := u.VerifyClaims(ctx, logCtx, c.Get("user").(*jwt.Token), sageUser)
	if !claims.IsVerified {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	username := claims.Username
	logCtx = logCtx.With("username", username)

	logCtx, project, err := getUUIDParam(c, logCtx, "_project")
	if err != nil {
		return err
	}

	req := sessions.Document{}
	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	if req.Name == "" {
		return e.ErrBad(logCtx, fid, "missing session name")
	}

	req.Status = "in progress"
	id, err := sessions.Post(ctx, logCtx, username, project, req)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to create session")
	}

	logCtx.Info("created session", fid)
	return c.JSON(http.StatusCreated, map[string]string{"session_id": id})
}

func getSessions(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.sage.getSessions")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	claims := u.VerifyClaims(ctx, logCtx, c.Get("user").(*jwt.Token), sageUser)
	if !claims.IsVerified {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	username := claims.Username
	logCtx = logCtx.With("username", username)

	logCtx, project, err := getUUIDParam(c, logCtx, "_project")
	if err != nil {
		return err
	}

	status := c.QueryParam("status")
	if status != "" {
		logCtx = logCtx.With("status", status)
	}

	sessions, err := sessions.GetAll(ctx, logCtx, username, project, status)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get projects")
	}

	logCtx.Info("get sessions", fid)
	return c.JSON(http.StatusOK, sessions)
}

func getSession(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.sage.getSession")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	claims := u.VerifyClaims(ctx, logCtx, c.Get("user").(*jwt.Token), sageUser)
	if !claims.IsVerified {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	username := claims.Username
	logCtx = logCtx.With("username", username)

	logCtx, project, err := getUUIDParam(c, logCtx, "_project")
	if err != nil {
		return err
	}

	logCtx, session, err := getUUIDParam(c, logCtx, "_session")
	if err != nil {
		return err
	}

	res, err := sessions.Get(ctx, logCtx, username, project, session)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get session")
	}

	logCtx.Info("get session", fid)
	return c.JSON(http.StatusOK, res)
}

func patchSession(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.sage.patchSession")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	claims := u.VerifyClaims(ctx, logCtx, c.Get("user").(*jwt.Token), sageUser)
	if !claims.IsVerified {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	username := claims.Username
	logCtx = logCtx.With("username", username)

	logCtx, project, err := getUUIDParam(c, logCtx, "_project")
	if err != nil {
		return err
	}

	logCtx, session, err := getUUIDParam(c, logCtx, "_session")
	if err != nil {
		return err
	}

	req := sessions.Document{}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	// Don't allow models to be patch by the UI.
	req.Models = nil

	if err := sessions.Patch(ctx, logCtx, username, project, session, req); err != nil {
		return e.Err(logCtx, err, fid, "unable to update session")
	}

	logCtx.Info("update session", fid)
	return c.NoContent(http.StatusOK)
}

func deleteSession(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.sage.deleteSession")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	claims := u.VerifyClaims(ctx, logCtx, c.Get("user").(*jwt.Token), sageUser)
	if !claims.IsVerified {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	username := claims.Username
	logCtx = logCtx.With("username", username)

	logCtx, project, err := getUUIDParam(c, logCtx, "_project")
	if err != nil {
		return err
	}

	logCtx, session, err := getUUIDParam(c, logCtx, "_session")
	if err != nil {
		return err
	}

	archive := c.QueryParam("archive") == "1"

	status := "deleted"
	if archive {
		status = "archived"
	}

	req := sessions.Document{Status: status}

	if err := sessions.Patch(ctx, logCtx, username, project, session, req); err != nil {
		return e.Err(logCtx, err, fid, "unable to update session")
	}

	logCtx.Info("update session", fid, "status", status)
	return c.NoContent(http.StatusOK)
}
