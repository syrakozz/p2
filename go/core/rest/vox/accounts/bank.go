package accounts

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"disruptive/pkg/vox/accounts"
	"disruptive/rest/auth"
	e "disruptive/rest/errors"
)

// GetBalance returns the user's bank balance info.
func GetBalance(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.accounts.GetBalance")

	b, err := accounts.GetBalance(ctx, logCtx)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get accounts bank details")
	}

	return c.JSON(http.StatusOK, b)
}

// GetAvailableBalance returns the user's available balances.
func GetAvailableBalance(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.accounts.GetAvailableBalance")

	b, err := accounts.GetAvailableBalance(ctx, logCtx)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get available balance")
	}

	return c.JSON(http.StatusOK, b)
}

// AddBalance adds the value to the user's balance.
func AddBalance(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.accounts.AddBalance")

	amount := strings.ToLower(c.QueryParam("amount"))

	if amount == "" {
		return e.ErrBad(logCtx, fid, "amount required")
	}

	a, err := strconv.Atoi(amount)
	if err != nil {
		return e.ErrBad(logCtx, fid, "invald amount")
	}

	res, err := accounts.AddBalance(ctx, logCtx, a)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to update accounts bank balance")
	}

	return c.JSON(http.StatusOK, res)
}

// ChargeBank charges an account's balance.
func ChargeBank(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.accounts.ChargeBank")

	characterVersion := strings.ToLower(c.QueryParam("character_version"))
	tier := c.QueryParam("tier")

	if characterVersion == "" {
		return e.ErrBad(logCtx, fid, "character_version required")
	}

	if tier == "" {
		return e.ErrBad(logCtx, fid, "tier required")
	}

	res, err := accounts.ChargeBank(ctx, logCtx, characterVersion, tier)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to update accounts bank balance")
	}

	return c.JSON(http.StatusOK, res)
}

// GetSubscription gets a user's subscription information.
func GetSubscription(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.accounts.GetSubscription")

	res, err := accounts.GetSubscription(ctx, logCtx)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get user's subscription information")
	}

	return c.JSON(http.StatusOK, res)
}

// PatchPendingSubscription sets the pending subscription value.
func PatchPendingSubscription(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.accounts.PendingSubscription")

	sku := strings.ToLower(c.QueryParam("sku"))

	if sku == "" {
		return e.ErrBad(logCtx, fid, "sku required")
	}

	res, err := accounts.PatchPendingSubscription(ctx, logCtx, sku)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to update accounts bank pending subscription")
	}

	return c.JSON(http.StatusOK, res)
}

// Subscribe subscribes a user to a new subscription model.
func Subscribe(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.accounts.Subscribe")

	sku := strings.ToLower(c.QueryParam("sku"))

	if sku == "" {
		return e.ErrBad(logCtx, fid, "sku required")
	}

	res, err := accounts.Subscribe(ctx, logCtx, sku)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to update accounts bank subscription")
	}

	return c.JSON(http.StatusOK, res)
}

// Unsubscribe unsubscribes a user.
func Unsubscribe(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.accounts.Unsubscribe")

	res, err := accounts.Unsubscribe(ctx, logCtx)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to unsubscribe accounts bank subscription")
	}

	return c.JSON(http.StatusOK, res)
}

// RedeemGiftCard redeems a gift card.
func RedeemGiftCard(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.accounts.RedeemGiftCard")

	giftCard := c.Param("gift_card")

	res, err := accounts.RedeemGiftCard(ctx, logCtx, giftCard)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to redeem gift card")
	}

	return c.JSON(http.StatusCreated, res)
}
