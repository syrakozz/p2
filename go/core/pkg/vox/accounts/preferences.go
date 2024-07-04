package accounts

import (
	"context"
	"log/slog"

	fs "cloud.google.com/go/firestore"

	"disruptive/lib/common"
	"disruptive/lib/firestore"
)

// GetAccountPreferences retrieves an account's preferences.
func GetAccountPreferences(ctx context.Context, logCtx *slog.Logger) (map[string]any, error) {
	fid := slog.String("fid", "vox.accounts.GetAccountPreferences")

	account := ctx.Value(common.AccountKey).(Document)

	collection := firestore.Client.Collection("accounts")
	if collection == nil {
		logCtx.Error("unable to get accounts collection", fid)
		return nil, common.ErrNotFound{}
	}

	doc, err := collection.Doc(account.ID).Get(ctx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to get accounts document", fid, "error", err)
		return nil, err
	}

	d := Document{}

	if err := doc.DataTo(&d); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to read account preferences data", fid, "error", err)
		return nil, err
	}

	return d.Preferences, nil
}

// PutAccountPreferences full updates a firestore account preferences.
func PutAccountPreferences(ctx context.Context, logCtx *slog.Logger, preferences map[string]any) error {
	fid := slog.String("fid", "vox.accounts.PutAccountPreferences")

	account := ctx.Value(common.AccountKey).(Document)

	collection := firestore.Client.Collection("accounts")
	if collection == nil {
		logCtx.Error("unable to get accounts collection", fid)
		return common.ErrNotFound{}
	}

	update := []fs.Update{
		{Path: "preferences", Value: preferences},
	}

	if _, err := collection.Doc(account.ID).Update(ctx, update); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to update account preferences", fid, "error", err)
		return err
	}

	return nil
}

// PatchAccountPreferences updates a firestore account preferences.
func PatchAccountPreferences(ctx context.Context, logCtx *slog.Logger, preferences map[string]any) (map[string]any, error) {
	fid := slog.String("fid", "vox.accounts.PatchAccountPreferences")

	account := ctx.Value(common.AccountKey).(Document)

	collection := firestore.Client.Collection("accounts")
	if collection == nil {
		logCtx.Error("unable to get accounts collection", fid)
		return nil, common.ErrNotFound{}
	}

	updates := make([]fs.Update, 0, len(preferences))
	for k, v := range preferences {
		updates = append(updates, fs.Update{Path: "preferences." + k, Value: v})
	}

	if len(updates) > 0 {
		if _, err := collection.Doc(account.ID).Update(ctx, updates); err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Error("unable to update account preferences", fid, "error", err)
			return nil, err
		}
	}

	prefs, err := GetAccountPreferences(ctx, logCtx)
	if err != nil {
		logCtx.Warn("unable to get account preferences", fid, "error", err)
		return nil, err
	}

	return prefs, nil
}

// DeleteAccountPreferences  deletes a firestore account preferences.
func DeleteAccountPreferences(ctx context.Context, logCtx *slog.Logger, preferences []string) (map[string]any, error) {
	fid := slog.String("fid", "vox.accounts.DeleteAccountPreferences")

	account := ctx.Value(common.AccountKey).(Document)

	collection := firestore.Client.Collection("accounts")
	if collection == nil {
		logCtx.Error("unable to get accounts collection", fid)
		return nil, common.ErrNotFound{}
	}

	updates := make([]fs.Update, 0, len(preferences))
	for _, v := range preferences {
		updates = append(updates, fs.Update{Path: "preferences." + v, Value: fs.Delete})
	}

	if _, err := collection.Doc(account.ID).Update(ctx, updates); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to delete account preferences", fid, "error", err)
		return nil, err
	}

	prefs, err := GetAccountPreferences(ctx, logCtx)
	if err != nil {
		logCtx.Warn("unable to get account preferences", fid, "error", err)
		return nil, err
	}

	return prefs, nil
}
