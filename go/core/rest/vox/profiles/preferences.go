package profiles

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"disruptive/pkg/vox/profiles"
	"disruptive/rest/auth"
	e "disruptive/rest/errors"
)

// GetPreferences return preferences for a profile.
func GetPreferences(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.profiles.GetPreferences")

	profileID := c.Param("profile_id")

	p, err := profiles.GetPreferences(ctx, logCtx, profileID)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get profile preferences")
	}

	return c.JSON(http.StatusOK, p)
}

// PutPreferences updates the entire profile preferences.
func PutPreferences(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.profiles.PutPreferences")

	profileID := c.Param("profile_id")

	p := map[string]any{}

	if err := (&echo.DefaultBinder{}).BindBody(c, &p); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	if err := profiles.PutPreferences(ctx, logCtx, profileID, p); err != nil {
		return e.Err(logCtx, err, fid, "unable to update profile preferences")
	}

	return c.NoContent(http.StatusOK)
}

// PatchPreferences update the profile preferences.
func PatchPreferences(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.profiles.PatchPreferences")

	profileID := c.Param("profile_id")

	p := map[string]any{}

	if err := (&echo.DefaultBinder{}).BindBody(c, &p); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	if len(p) < 1 {
		return e.ErrBad(logCtx, fid, "invalid patch object")
	}

	res, err := profiles.PatchPreferences(ctx, logCtx, profileID, p)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to update profile preferences")
	}

	return c.JSON(http.StatusOK, res)
}

// DeletePreferences delete profile preferences.
func DeletePreferences(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.profiles.DeletePreferences")

	profileID := c.Param("profile_id")

	p := []string{}

	if err := (&echo.DefaultBinder{}).BindBody(c, &p); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	if len(p) < 1 {
		return e.ErrBad(logCtx, fid, "invalid list of keys")
	}

	res, err := profiles.DeletePreferences(ctx, logCtx, profileID, p)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to delete profile preferences")
	}

	return c.JSON(http.StatusOK, res)
}
