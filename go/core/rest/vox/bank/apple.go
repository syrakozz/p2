package bank

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"disruptive/lib/common"
	"disruptive/pkg/vox/accounts"
	"disruptive/pkg/vox/bank"
	"disruptive/rest/auth"
	e "disruptive/rest/errors"
)

// PostAppleIAPTransaction is the REST API for creating Apple IAP transactions.
func PostAppleIAPTransaction(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.bank.PostAppleIAPTransaction")
	account := ctx.Value(common.AccountKey).(accounts.Document)

	req := bank.AppleStoreIAPTransaction{}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read apple iap transaction")
	}

	if req.ProductID == "" {
		return e.ErrBad(logCtx, fid, "missing productId")
	}

	if req.TransactionID == "" {
		return e.ErrBad(logCtx, fid, "missing transactionId")
	}

	if req.TransactionReceipt == "" {
		return e.ErrBad(logCtx, fid, "missing transactionReceipt")
	}

	txn, err := bank.PostAppleIAPTransaction(ctx, logCtx, account.ID, req)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to create apple iap transaction")
	}

	return c.JSON(http.StatusCreated, txn)
}
