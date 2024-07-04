package accounts

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"disruptive/pkg/vox/accounts"
	"disruptive/rest/auth"
	e "disruptive/rest/errors"
)

// GetPreferences returns account preferences.
func GetPreferences(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.accounts.GetPreferences")

	p, err := accounts.GetAccountPreferences(ctx, logCtx)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get account preferences")
	}

	return c.JSON(http.StatusOK, p)
}

// PutPreferences update the entire account preferences.
func PutPreferences(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.accounts.PutPreferences")

	p := map[string]any{}

	if err := c.Bind(&p); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	if err := accounts.PutAccountPreferences(ctx, logCtx, p); err != nil {
		return e.Err(logCtx, err, fid, "unable to update account preferences")
	}

	return c.NoContent(http.StatusOK)
}

// PatchPreferences update accounts preferences.
func PatchPreferences(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.accounts.PatchPreferences")

	p := map[string]any{}

	if err := c.Bind(&p); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	if len(p) < 1 {
		return e.ErrBad(logCtx, fid, "invalid patch object")
	}

	res, err := accounts.PatchAccountPreferences(ctx, logCtx, p)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to update account preferences")
	}

	return c.JSON(http.StatusOK, res)
}

// DeletePreferences deletet accounts preferences.
func DeletePreferences(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.accounts.DeletePreferences")

	p := []string{}

	if err := c.Bind(&p); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	if len(p) < 1 {
		return e.ErrBad(logCtx, fid, "invalid list of keys")
	}

	res, err := accounts.DeleteAccountPreferences(ctx, logCtx, p)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to delete account preferences")
	}

	return c.JSON(http.StatusOK, res)
}
