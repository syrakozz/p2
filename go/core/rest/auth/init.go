// Package auth functions
package auth

import (
	"context"
	"log/slog"

	"github.com/labstack/echo/v4"

	"disruptive/lib/common"
	"disruptive/pkg/vox/accounts"
)

// InitRequest initializes REST functions
func InitRequest(c echo.Context, name string) (context.Context, *slog.Logger, slog.Attr) {
	fid := slog.String("fid", name)
	ctx := c.Request().Context()
	account := ctx.Value(common.AccountKey).(accounts.Document)
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID), "uid", account.ID)

	return ctx, logCtx, fid
}
