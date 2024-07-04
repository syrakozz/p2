// Package demo ...
package demo

import (
	"context"
	"log/slog"

	fs "cloud.google.com/go/firestore"

	"disruptive/lib/common"
	"disruptive/lib/firestore"
)

// Document contains the demo user fields.
type Document struct {
	Status *string `firestore:"status" json:"status"`
}

// PatchUser makes a document inactive
func PatchUser(ctx context.Context, logCtx *slog.Logger, user string, document Document) error {
	fid := slog.String("fid", "vox.demo.Patch")

	collection := firestore.Client.Collection("demo")
	if collection == nil {
		logCtx.Error("demo collection not found", fid)
		return common.ErrNotFound{}
	}

	update := []fs.Update{}

	if document.Status != nil {
		update = append(update, fs.Update{Path: "status", Value: *document.Status})
	}

	if len(update) < 1 {
		logCtx.Warn("nothing to update")
		return nil
	}

	if _, err := collection.Doc(user).Update(ctx, update); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to update demo user document", fid, "error", err)
		return err
	}

	return nil
}
