package bank

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"disruptive/lib/configs"
	"disruptive/rest/auth"
	e "disruptive/rest/errors"
)

// GetRates returns rate configs.
func GetRates(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.bank.GetRates")

	r, err := configs.GetRates(ctx, logCtx)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get accounts bank rates")
	}

	return c.JSON(http.StatusOK, r)
}

// GetSKUs returns SKU configs.
func GetSKUs(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.bank.GetSKUs")

	s, err := configs.GetSKUs(ctx, logCtx)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get SKU config details")
	}

	return c.JSON(http.StatusOK, s)
}
