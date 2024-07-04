package accounts

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"disruptive/lib/common"
	"disruptive/pkg/vox/accounts"
	"disruptive/rest/auth"
	e "disruptive/rest/errors"
)

// GetAccount returns the user's account.
func GetAccount(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.accounts.GetAccount")

	uid := c.Param("account_id")
	logCtx = logCtx.With("for_uid", uid)

	p, err := accounts.GetAccount(ctx, logCtx, uid)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get account")
	}

	return c.JSON(http.StatusOK, p)
}

// GetAccountMe returns the user's account.
func GetAccountMe(c echo.Context) error {
	ctx := c.Request().Context()
	account := ctx.Value(common.AccountKey).(accounts.Document)
	return c.JSON(http.StatusOK, account)
}

// PatchAccountMe patches an account's basic information.
func PatchAccountMe(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.accounts.PatchAccountMe")
	account := ctx.Value(common.AccountKey).(accounts.Document)

	p := accounts.PatchDocument{}

	if err := c.Bind(&p); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	p.ID = account.ID

	a, err := accounts.PatchAccount(ctx, logCtx, p)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to patch account")
	}

	return c.JSON(http.StatusOK, a)
}

// DeleteAccountMe sets an account to inactive or full deletes a profile.
func DeleteAccountMe(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.accounts.DeleteAccountMe")
	account := ctx.Value(common.AccountKey).(accounts.Document)

	if err := accounts.DeleteAccount(ctx, logCtx, account.ID); err != nil {
		return e.Err(logCtx, err, fid, "unable to set delete account")
	}

	return c.NoContent(http.StatusNoContent)
}
