package bank

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"disruptive/pkg/vox/bank"
	"disruptive/rest/auth"
	e "disruptive/rest/errors"
)

// GetGiftCard gets a current gift cards.
func GetGiftCard(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.bank.GetGiftCard")

	giftCard := c.Param("gift_card")

	res, err := bank.GetGiftCard(ctx, logCtx, giftCard)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get gift card")
	}

	return c.JSON(http.StatusOK, res)
}

// GetGiftCards gets current or expired gift cards.
func GetGiftCards(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.bank.GetGiftCards")

	expired := c.QueryParam("expired")

	res, err := bank.GetGiftCards(ctx, logCtx, expired)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get gift cards")
	}

	return c.JSON(http.StatusOK, res)
}

// ExpireGiftCard expires a gift card.
func ExpireGiftCard(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.bank.ExpireGiftCard")

	giftCard := c.Param("gift_card")

	if giftCard == "" {
		return e.ErrBad(logCtx, fid, "gift card required")
	}

	err := bank.ExpireGiftCard(ctx, logCtx, giftCard)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to delete gift card")
	}

	return c.NoContent(http.StatusOK)
}
