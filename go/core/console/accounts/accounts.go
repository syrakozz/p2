// Package accounts contains console commands for accounts.
package accounts

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"disruptive/lib/common"
	"disruptive/pkg/vox/accounts"
)

// Get retrieves an account.
func Get(ctx context.Context, searchValue string) error {
	var (
		account accounts.Document
		err     error
		logCtx  *slog.Logger
	)

	if strings.Contains(searchValue, "@") {
		logCtx = slog.With("email", searchValue)
		account, err = accounts.GetAccountByEmail(ctx, logCtx, searchValue)
	} else {
		logCtx = slog.With("account_id", searchValue)
		account, err = accounts.GetAccount(ctx, logCtx, searchValue)
	}

	if err != nil {
		if errors.Is(err, common.ErrNotFound{}) {
			logCtx.Info("account not found")
			return nil
		}

		logCtx.Error("unable to get account", "error", err)
		return err
	}

	common.P(account)
	return nil
}

// PatchInactive patches an account's inactive field.
func PatchInactive(ctx context.Context, id string, inactive bool) error {
	logCtx := slog.With("account_id", id)
	ctx = context.WithValue(ctx, common.AccountKey, accounts.Document{ID: id})

	account, err := accounts.PatchAccount(ctx, logCtx, accounts.PatchDocument{ID: id, Inactive: &inactive})
	if err != nil {
		if errors.Is(err, common.ErrNotFound{}) {
			logCtx.Info("account not found")
			return nil
		}

		logCtx.Error("unable to patch account inactive field", "error", err)
		return err
	}

	common.P(account)
	return nil
}

// Delete deletes an account if force is true, otherwise sets inactive to true.
func Delete(ctx context.Context, accountID string, force bool) error {
	logCtx := slog.With("account_id", accountID, "force", force)

	ctx = context.WithValue(ctx, common.AccountKey, accounts.Document{ID: accountID})

	inactive := true
	if _, err := accounts.PatchAccount(ctx, logCtx, accounts.PatchDocument{ID: accountID, Inactive: &inactive}); err != nil {
		if errors.Is(err, common.ErrNotFound{}) {
			logCtx.Info("account not found")
			return nil
		}

		logCtx.Error("unable to patch account inactive field", "error", err)
		return err
	}

	return nil
}
