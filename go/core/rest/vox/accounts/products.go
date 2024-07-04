// Package accounts handles account APIs.
package accounts

import (
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"disruptive/lib/common"
	"disruptive/lib/configs"
	"disruptive/pkg/vox/accounts"
	"disruptive/rest/auth"
	e "disruptive/rest/errors"
)

// PostAccountMeProduct adds a character to an account.
func PostAccountMeProduct(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.accounts.PostAccountMeProduct")
	account := ctx.Value(common.AccountKey).(accounts.Document)

	product := c.Param("product")

	if err := accounts.PutProducts(ctx, logCtx, account.ID, product, accounts.Product{}); err != nil {
		return e.Err(logCtx, err, fid, "unable to add product to account")
	}

	return c.NoContent(http.StatusCreated)
}

// PostAccountMeProductConnect connects a character device to an account.
func PostAccountMeProductConnect(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.accounts.PostAccountMeProductConnect")

	product := c.Param("product")
	deviceID := c.Param("device_id")

	cfg, err := configs.Get(ctx, logCtx, "products")
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get config")
	}

	productCfg, ok := cfg[product].(map[string]any)
	if !ok {
		return e.Err(logCtx, err, fid, "unable to get product config")
	}

	whiteList, ok := productCfg["white_list"].(bool)
	if !ok {
		logCtx.Info("white_list config value does not exist, defaulting to false")
		whiteList = false
	}

	id := deviceID
	if !whiteList && !strings.Contains(deviceID, "-") {
		d, err := strconv.ParseUint(strings.ReplaceAll(deviceID, ":", ""), 16, 64)
		if err != nil {
			return e.ErrBad(logCtx, fid, "invalid MAC address")
		}

		mac, err := net.ParseMAC(common.IntToMACAddress(d))
		if err != nil {
			return e.ErrBad(logCtx, fid, "invalid MAC address")
		}

		id = mac.String()
	}

	res, err := accounts.ConnectProductDevice(ctx, logCtx, product, id)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to connect product to account")
	}

	if res.FirstTime {
		return c.JSON(http.StatusCreated, res)
	}
	return c.JSON(http.StatusOK, res)
}

// DeleteAccountMeProduct removes a character from an account.
func DeleteAccountMeProduct(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.accounts.DeleteAccountMeProduct")
	account := ctx.Value(common.AccountKey).(accounts.Document)

	product := c.Param("product")

	if err := accounts.DeleteCharacter(ctx, logCtx, account.ID, product); err != nil {
		return e.Err(logCtx, err, fid, "unable to delete product from account")
	}

	return c.NoContent(http.StatusOK)
}
