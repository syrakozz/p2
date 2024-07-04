// Package profiles ...
package profiles

import (
	"context"
	"fmt"
	"log/slog"

	fs "cloud.google.com/go/firestore"

	"disruptive/lib/common"
	"disruptive/lib/firestore"
	"disruptive/pkg/vox/accounts"
)

// GetPreferences retrieves an profile's preferences.
func GetPreferences(ctx context.Context, logCtx *slog.Logger, profileID string) (map[string]any, error) {
	fid := slog.String("fid", "vox.profiles.GetPreferences")

	account := ctx.Value(common.AccountKey).(accounts.Document)

	path := fmt.Sprintf("accounts/%s/profiles", account.ID)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("profiles collection not found", fid)
		return nil, common.ErrNotFound{}
	}

	doc, err := collection.Doc(profileID).Get(ctx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to get profiles document", fid, "error", err)
		return nil, err
	}

	d := Document{}

	if err := doc.DataTo(&d); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to read profiles preferences data", fid, "error", err)
		return nil, err
	}

	return d.Preferences, nil
}

// PutPreferences full updates a firestore profile preferences.
func PutPreferences(ctx context.Context, logCtx *slog.Logger, profileID string, preferences map[string]any) error {
	fid := slog.String("fid", "vox.profiles.PutPreferences")

	account := ctx.Value(common.AccountKey).(accounts.Document)

	path := fmt.Sprintf("accounts/%s/profiles", account.ID)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("profiles collection not found", fid)
		return common.ErrNotFound{}
	}

	update := []fs.Update{
		{Path: "preferences", Value: preferences},
	}

	if _, err := collection.Doc(profileID).Update(ctx, update); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to update profiles preferences", fid, "error", err)
		return err
	}

	return nil
}

// PatchPreferences updates a firestore profile preferences.
func PatchPreferences(ctx context.Context, logCtx *slog.Logger, profileID string, preferences map[string]any) (map[string]any, error) {
	fid := slog.String("fid", "vox.profiles.PatchPreferences")

	account := ctx.Value(common.AccountKey).(accounts.Document)

	path := fmt.Sprintf("accounts/%s/profiles", account.ID)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("profiles collection not fonud", fid)
		return nil, common.ErrNotFound{}
	}

	updates := make([]fs.Update, 0, len(preferences))
	for k, v := range preferences {
		updates = append(updates, fs.Update{Path: "preferences." + k, Value: v})
	}

	if len(updates) > 0 {
		if _, err := collection.Doc(profileID).Update(ctx, updates); err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Error("unable to update profiles preferences", fid, "error", err)
			return nil, err
		}
	}

	prefs, err := GetPreferences(ctx, logCtx, profileID)
	if err != nil {
		logCtx.Warn("unable to get profile preferences", fid, "error", err)
		return nil, err
	}

	return prefs, nil
}

// DeletePreferences  deletes a firestore profile preferences.
func DeletePreferences(ctx context.Context, logCtx *slog.Logger, profileID string, preferences []string) (map[string]any, error) {
	fid := slog.String("fid", "vox.profiles.DeleteProfilePreferences")

	account := ctx.Value(common.AccountKey).(accounts.Document)

	path := fmt.Sprintf("accounts/%s/profiles", account.ID)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("profiles collection not found", fid)
		return nil, common.ErrNotFound{}
	}

	updates := make([]fs.Update, 0, len(preferences))
	for _, v := range preferences {
		updates = append(updates, fs.Update{Path: "preferences." + v, Value: fs.Delete})
	}

	if _, err := collection.Doc(profileID).Update(ctx, updates); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to delete profiles preferences", fid, "error", err)
		return nil, err
	}

	prefs, err := GetPreferences(ctx, logCtx, profileID)
	if err != nil {
		logCtx.Warn("unable to get profile preferences", fid, "error", err)
		return nil, err
	}

	return prefs, nil
}
