package notifications

import (
	"context"
	"fmt"
	"log/slog"

	fs "cloud.google.com/go/firestore"

	"disruptive/lib/common"
	"disruptive/lib/firestore"
	"disruptive/pkg/vox/accounts"
)

// PatchDocument contains a notification document.
type PatchDocument struct {
	ID       string `json:"id"`
	Inactive *bool  `json:"inactive"`
	Read     *bool  `json:"read"`
}

// GetNotifications returns all notifications for the account.
func GetNotifications(ctx context.Context, logCtx *slog.Logger, all, inactive bool) ([]Document, error) {
	fid := slog.String("fid", "vox.notifications.GetNotifications")

	account := ctx.Value(common.AccountKey).(accounts.Document)

	path := fmt.Sprintf("accounts/%s/notifications", account.ID)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("notifications collection not found", fid)
		return nil, common.ErrNotFound{}
	}

	documents := []Document{}

	var iter *fs.DocumentIterator
	if all {
		iter = collection.Documents(ctx)
	} else if inactive {
		iter = collection.Where("inactive", "==", true).Documents(ctx)
	} else {
		iter = collection.Where("inactive", "!=", true).Documents(ctx)
	}

	docs, err := iter.GetAll()
	if err != nil {
		logCtx.Error("unable to get documents", fid, "error", err)
		return nil, err
	}

	for _, doc := range docs {
		d := Document{}
		if err := doc.DataTo(&d); err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Error("unable to read notification data", fid, "error", err)
			return nil, err
		}

		documents = append(documents, d)
	}

	return documents, nil
}

// GetNotification modified an account notification and returns the modified document.
func GetNotification(ctx context.Context, logCtx *slog.Logger, id string) (Document, error) {
	fid := slog.String("fid", "vox.notifications.GetNotification")

	account := ctx.Value(common.AccountKey).(accounts.Document)

	d := Document{}

	path := fmt.Sprintf("accounts/%s/notifications", account.ID)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Warn("notifications collection not found", fid)
		return d, common.ErrNotFound{}
	}

	doc, err := collection.Doc(id).Get(ctx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to get notification document", fid, "error", err)
		return d, err
	}

	if err := doc.DataTo(&d); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to read notification data", fid, "error", err)
		return d, err
	}

	return d, nil
}

// PatchNotification modified an account notification and returns the modified document.
func PatchNotification(ctx context.Context, logCtx *slog.Logger, document PatchDocument) (Document, error) {
	fid := slog.String("fid", "vox.notifications.PatchNotification")

	account := ctx.Value(common.AccountKey).(accounts.Document)

	path := fmt.Sprintf("accounts/%s/notifications", account.ID)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Warn("notifications collection not found", fid)
		return Document{}, common.ErrNotFound{}
	}

	update := []fs.Update{}

	if document.Read != nil {
		update = append(update, fs.Update{Path: "read", Value: *document.Read})
	}

	if document.Inactive != nil {
		update = append(update, fs.Update{Path: "inactive", Value: *document.Inactive})
	}

	if len(update) > 0 {
		if _, err := collection.Doc(document.ID).Update(ctx, update); err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Warn("unable to update notification document", fid, "error", err)
			return Document{}, err
		}
	}

	d, err := GetNotification(ctx, logCtx, document.ID)
	if err != nil {
		logCtx.Warn("unable to get notification document", fid, "error", err)
		return Document{}, err
	}

	return d, nil
}

// DeleteNotification deletes an account notification.
func DeleteNotification(ctx context.Context, logCtx *slog.Logger, id string) error {
	fid := slog.String("fid", "vox.notifications.DeleteNotification")

	account := ctx.Value(common.AccountKey).(accounts.Document)

	path := fmt.Sprintf("accounts/%s/notifications", account.ID)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Warn("notifications collection not found", fid)
		return common.ErrNotFound{}
	}

	if _, err := collection.Doc(id).Delete(ctx); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to delete notification document", fid, "error", err)
		return err
	}

	return nil
}
