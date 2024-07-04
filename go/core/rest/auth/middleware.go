package auth

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"firebase.google.com/go/auth"
	"github.com/labstack/echo/v4"

	"disruptive/lib/common"
	"disruptive/lib/firebase"
	"disruptive/pkg/vox/accounts"
	e "disruptive/rest/errors"
)

// SetMiddleware function to set the me variable
func SetMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	fid := slog.String("fid", "auth.SetMiddleware")
	logCtx := slog.With()

	return func(c echo.Context) error {
		ctx := c.Request().Context()
		header := c.Request().Header

		var bearer string

		bearerList := strings.Split(header.Get("Authorization"), "Bearer ")
		if len(bearerList) == 2 {
			bearer = bearerList[1]
		} else {
			bearer = c.QueryParam("authorization")
		}

		if bearer != "" {
			fbuser, err := firebase.GetUserByJWT(ctx, logCtx, bearer)
			if err != nil {
				return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
			}

			account, err := getOrCreateAccount(ctx, logCtx, fbuser)
			if err != nil {
				return err
			}

			if account.Inactive {
				return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
			}

			ctx = context.WithValue(ctx, common.FirebaseUserKey, fbuser)
			ctx = context.WithValue(ctx, common.AccountKey, account)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}

		uid := header.Get("X-UID")
		if uid != "" {
			fbuser, err := firebase.GetUserByUID(ctx, logCtx, uid)
			if err != nil {
				return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
			}

			account, err := getOrCreateAccount(ctx, logCtx, fbuser)
			if err != nil {
				return err
			}

			if account.Inactive {
				return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
			}

			ctx = context.WithValue(ctx, common.FirebaseUserKey, fbuser)
			ctx = context.WithValue(ctx, common.AccountKey, account)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}

		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}
}

// SetAdminMiddleware function to set the me variable
func SetAdminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	fid := slog.String("fid", "auth.SetAdminMiddleware")
	logCtx := slog.With()

	return func(c echo.Context) error {
		ctx := c.Request().Context()
		header := c.Request().Header

		bearer := strings.Split(header.Get("Authorization"), "Bearer ")
		if len(bearer) == 2 {
			fbuser, err := firebase.GetUserByJWT(ctx, logCtx, bearer[1])
			if err != nil {
				return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
			}

			account, err := getOrCreateAccount(ctx, logCtx, fbuser)
			if err != nil {
				return err
			}

			if !account.Admin || account.Inactive {
				return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
			}

			ctx = context.WithValue(ctx, common.FirebaseUserKey, fbuser)
			ctx = context.WithValue(ctx, common.AccountKey, account)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}

		uid := header.Get("X-UID")
		if uid != "" {
			fbuser, err := firebase.GetUserByUID(ctx, logCtx, uid)
			if err != nil {
				return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
			}

			account, err := getOrCreateAccount(ctx, logCtx, fbuser)
			if err != nil {
				return err
			}

			if !account.Admin || account.Inactive {
				return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
			}

			ctx = context.WithValue(ctx, common.FirebaseUserKey, fbuser)
			ctx = context.WithValue(ctx, common.AccountKey, account)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}

		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}
}

func getOrCreateAccount(ctx context.Context, logCtx *slog.Logger, fbuser *auth.UserRecord) (accounts.Document, error) {
	fid := slog.String("fid", "auth.getOrCreateAccount")

	account, err := accounts.GetAccount(ctx, logCtx, fbuser.UID)
	if err != nil {
		if errors.Is(err, common.ErrNotFound{}) {
			d := accounts.Document{
				ID:          fbuser.UID,
				DisplayName: fbuser.DisplayName,
				Email:       fbuser.Email,
			}

			account, err = accounts.CreateAccount(ctx, logCtx, d)
			if err != nil {
				return accounts.Document{}, e.Err(logCtx, err, fid, "unable to create account")
			}

			return account, nil
		}
		return accounts.Document{}, e.Err(logCtx, err, fid, "unable to get account")
	}

	return account, nil
}
