package sage

import (
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"

	"disruptive/lib/common"
	"disruptive/lib/sage/projects"
	u "disruptive/lib/users"
	e "disruptive/rest/errors"
)

func postProjects(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.sage.postProjects")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	claims := u.VerifyClaims(ctx, logCtx, c.Get("user").(*jwt.Token), sageUser)
	if !claims.IsVerified {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	username := claims.Username
	logCtx = logCtx.With("username", username)

	req := projects.Document{}
	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	if req.Name == "" {
		return e.ErrBad(logCtx, fid, "missing project name")
	}

	req.Status = "in progress"
	id, err := projects.Post(ctx, logCtx, username, req)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to create project")
	}

	logCtx.Info("created project", fid)
	return c.JSON(http.StatusCreated, map[string]string{"project_id": id})
}

func getProjects(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.sage.getProjects")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	claims := u.VerifyClaims(ctx, logCtx, c.Get("user").(*jwt.Token), sageUser)
	if !claims.IsVerified {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	username := claims.Username
	logCtx = logCtx.With("username", username)

	status := c.QueryParam("status")
	if status != "" {
		logCtx = logCtx.With("status", status)
	}

	projects, err := projects.GetAll(ctx, logCtx, username, status)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get projects")
	}

	logCtx.Info("get projects", fid)
	return c.JSON(http.StatusOK, projects)
}

func getProject(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.sage.getProject")
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

	res, err := projects.Get(ctx, logCtx, username, project)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get project")
	}

	logCtx.Info("get projects", fid)
	return c.JSON(http.StatusOK, res)
}

func patchProject(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.sage.patchProject")
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

	req := projects.Document{}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	if err := projects.Patch(ctx, logCtx, username, project, req); err != nil {
		return e.Err(logCtx, err, fid, "unable to update project")
	}

	logCtx.Info("update project", fid)
	return c.NoContent(http.StatusOK)
}

func deleteProject(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.sage.deleteProject")
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

	archive := c.QueryParam("archive") == "1"

	status := "deleted"
	if archive {
		status = "archived"
	}

	req := projects.Document{Status: status}

	if err := projects.Patch(ctx, logCtx, username, project, req); err != nil {
		return e.Err(logCtx, err, fid, "unable to update project")
	}

	logCtx.Info("update project", fid, "status", status)
	return c.NoContent(http.StatusOK)
}
