package sage

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"

	"disruptive/lib/common"
	"disruptive/lib/sage/models"
	u "disruptive/lib/users"
	e "disruptive/rest/errors"
)

func postModels(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.sage.postModels")
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

	req := models.Document{}
	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	if req.Type == "" {
		return e.ErrBad(logCtx, fid, "invalid type")
	}

	if req.Model == "" {
		return e.ErrBad(logCtx, fid, "invalid model")
	}

	if req.Personality != "" && req.PersonalityKey != "" {
		return e.ErrBad(logCtx, fid, "personality and personality_key may not both be set")
	}

	id, err := models.Post(ctx, logCtx, username, project, session, req)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to create model")
	}

	logCtx.Info("created model")
	return c.JSON(http.StatusCreated, map[string]string{"id": id})
}

func getModel(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.sage.getModel")
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

	logCtx, model, err := getUUIDParam(c, logCtx, "_model")
	if err != nil {
		return err
	}

	res, err := models.Get(ctx, logCtx, username, project, session, model)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get model")
	}

	logCtx.Info("get model", fid)
	return c.JSON(http.StatusOK, res)
}

func patchModel(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.sage.patchModel")
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

	logCtx, model, err := getUUIDParam(c, logCtx, "_model")
	if err != nil {
		return err
	}

	req := models.Document{}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	if req.Personality != "" && req.PersonalityKey != "" {
		return e.ErrBad(logCtx, fid, "personality and personality_key may not both be set")
	}

	if err := models.Patch(ctx, logCtx, username, project, session, model, req); err != nil {
		return e.Err(logCtx, err, fid, "unable to update model")
	}

	logCtx.Info("update model", fid)
	return c.NoContent(http.StatusOK)
}

func postModelEntry(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.sage.postModelEntry")
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

	logCtx, model, err := getUUIDParam(c, logCtx, "_model")
	if err != nil {
		return err
	}

	entry := c.Param("_entry")
	if entry == "" {
		return e.ErrBad(logCtx, fid, "invalid entry")
	}
	logCtx = logCtx.With("entry", entry)

	req := struct {
		Text string `json:"text"`
	}{}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	if req.Text == "" {
		return e.ErrBad(logCtx, fid, "invalid to data")
	}

	res, err := models.PostEntry(ctx, logCtx, username, project, session, model, entry, req.Text)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to post model entry")
	}

	logCtx.Info("create new version", fid)
	return c.JSON(http.StatusOK, res)
}

func getModelEntry(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.sage.getModelEntry")
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

	logCtx, model, err := getUUIDParam(c, logCtx, "_model")
	if err != nil {
		return err
	}

	entry := c.Param("_entry")
	if entry == "" {
		return e.ErrBad(logCtx, fid, "invalid entry")
	}
	logCtx = logCtx.With("entry", entry)

	res, err := models.GetEntry(ctx, logCtx, username, project, session, model, entry)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get model entry")
	}

	logCtx.Info("get model entry", fid)
	return c.JSON(http.StatusOK, res)
}

func postModelEntryVersion(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.sage.postModelEntryVersion")
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

	logCtx, model, err := getUUIDParam(c, logCtx, "_model")
	if err != nil {
		return err
	}

	entry := c.Param("_entry")
	if entry == "" {
		return e.ErrBad(logCtx, fid, "invalid entry")
	}
	logCtx = logCtx.With("entry", entry)

	version := c.Param("_version")
	if version == "" {
		return e.ErrBad(logCtx, fid, "invalid entry")
	}
	logCtx = logCtx.With("entry", entry)

	words := c.QueryParam("words")
	if words != "" {
		if _, err := strconv.Atoi(words); err != nil {
			return e.ErrBad(logCtx, fid, "invalid words number")
		}
	}

	bullets := c.QueryParam("bullets")
	if bullets != "" {
		if _, err := strconv.Atoi(bullets); err != nil {
			return e.ErrBad(logCtx, fid, "invalid bullets number")
		}
	}

	if words == "" && bullets == "" {
		return e.ErrBad(logCtx, fid, "missing words or bullets")
	}

	res, err := models.PostEntryVersion(ctx, logCtx, username, project, session, model, entry, version, words, bullets)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to put model current version")
	}

	logCtx.Info("updated model current version", fid)
	return c.JSON(http.StatusOK, res)

}

func putModelEntryVersion(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.sage.putModelEntryVersion")
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

	logCtx, model, err := getUUIDParam(c, logCtx, "_model")
	if err != nil {
		return err
	}

	entry := c.Param("_entry")
	if entry == "" {
		return e.ErrBad(logCtx, fid, "invalid entry")
	}
	logCtx = logCtx.With("entry", entry)

	version := c.Param("_version")
	if version == "" {
		return e.ErrBad(logCtx, fid, "invalid entry")
	}
	logCtx = logCtx.With("entry", entry)

	res, err := models.PutEntryVersion(ctx, logCtx, username, project, session, model, entry, version)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to put model current version")
	}

	logCtx.Info("updated model current version", fid)
	return c.JSON(http.StatusOK, res)
}
