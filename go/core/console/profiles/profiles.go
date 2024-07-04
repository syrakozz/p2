// Package profiles contains console functions for profiles.
package profiles

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"disruptive/lib/common"
	"disruptive/pkg/vox/accounts"
	"disruptive/pkg/vox/profiles"
)

// GetIDs retrieves all profile IDs for an account by accountID or email.
func GetIDs(ctx context.Context, accountID string) error {
	logCtx := slog.With("account_id", accountID)

	ids, err := profiles.GetIDsByAccount(ctx, logCtx, accountID)
	if err != nil {
		if errors.Is(err, common.ErrNotFound{}) {
			logCtx.Info("profile not found")
			return err
		}

		logCtx.Error("unable to get profile", "error", err)
		return err
	}

	fmt.Println("account_id:", accountID)
	fmt.Println()
	for _, id := range ids {
		fmt.Println("profile_id:", id)
	}

	return nil
}

// Get retrieves a profile.
func Get(ctx context.Context, accountID, profileID string) error {
	logCtx := slog.With("account_id", accountID, "profile_id", profileID)

	ctx = context.WithValue(ctx, common.AccountKey, accounts.Document{ID: accountID})

	p, err := profiles.GetByID(ctx, logCtx, profileID)
	if err != nil {
		if errors.Is(err, common.ErrNotFound{}) {
			logCtx.Info("profile not found")
			return err
		}

		logCtx.Error("unable to get profile", "error", err)
		return err
	}

	common.P(p)
	return nil
}

// PatchInactive patches a profile's inactive value.
func PatchInactive(ctx context.Context, accountID, profileID string, inactive bool) error {
	logCtx := slog.With("account_id", accountID, "profile_id", profileID, "inactive", inactive)

	ctx = context.WithValue(ctx, common.AccountKey, accounts.Document{ID: accountID})

	p, err := profiles.Patch(ctx, logCtx, profileID, profiles.PatchDocument{Inactive: &inactive})
	if err != nil {
		if errors.Is(err, common.ErrNotFound{}) {
			logCtx.Info("profile not found")
			return err
		}

		logCtx.Error("unable to patch profile inactive", "error", err)
		return err
	}

	common.P(p)
	return nil
}

// Delete deletes a profile if force is true, otherwise sets inactive to true.
func Delete(ctx context.Context, accountID, profileID string, force bool) error {
	logCtx := slog.With("account_id", accountID, "profile_id", profileID, "force", force)

	ctx = context.WithValue(ctx, common.AccountKey, accounts.Document{ID: accountID})

	inactive := true
	if _, err := profiles.Patch(ctx, logCtx, profileID, profiles.PatchDocument{Inactive: &inactive}); err != nil {
		if errors.Is(err, common.ErrNotFound{}) {
			logCtx.Info("profile not found")
			return err
		}

		logCtx.Error("unable to patch profile inactive", "error", err)
		return err
	}

	return nil
}
