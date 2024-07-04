package profiles

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"path"
	"path/filepath"

	"disruptive/lib/common"
	"disruptive/lib/firebase"
	"disruptive/lib/firestore"
	"disruptive/lib/gcp"
	"disruptive/pkg/vox/accounts"

	fs "cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
)

// GetProfilePicture retrieves a profile's picture.
func GetProfilePicture(ctx context.Context, logCtx *slog.Logger, profileID string) (io.ReadCloser, string, error) {
	fid := slog.String("fid", "vox.profiles.GetProfilePicture")

	profile, err := GetByID(ctx, logCtx, profileID)
	if err != nil {
		logCtx.Error("unable to get profile", fid, "error", err)
		return nil, "", err
	}

	if profile.Picture == "" {
		logCtx.Error("profile picture not found", fid, "error", err)
		return nil, "", common.ErrNotFound{Msg: "profile picture not found"}
	}

	rc, cType, err := gcp.Storage.Download(ctx, firebase.GCSBucket, profile.Picture)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			logCtx.Error("unable to download profile picture", fid, "error", err)
			return nil, "", common.ErrGone{Msg: "profile picture gone"}
		}

		logCtx.Error("unable to download profile picture file", fid, "error", err)
		return nil, "", err
	}

	return rc, cType, nil
}

// PutProfilePicture replaces a picture in GCS and the path in firestore.
func PutProfilePicture(ctx context.Context, logCtx *slog.Logger, profileID, filename string, r io.Reader) error {
	fid := slog.String("fid", "vox.profiles.PutProfilePicture")

	account := ctx.Value(common.AccountKey).(accounts.Document)

	ext := path.Ext(filename)
	contentType := mime.TypeByExtension(ext)
	gcsPath := filepath.Join("store", "accounts", account.ID, "profiles", profileID, "picture"+ext)

	if err := gcp.Storage.Upload(ctx, r, firebase.GCSBucket, gcsPath, contentType); err != nil {
		logCtx.Error("unable to upload profile picture", fid, "error", err)
		return err
	}

	path := fmt.Sprintf("accounts/%s/profiles", account.ID)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("profiles collection not found", fid)
		return common.ErrNotFound{}
	}

	update := []fs.Update{{Path: "picture", Value: gcsPath}}

	if _, err := collection.Doc(profileID).Update(ctx, update); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to update profile document", fid, "error", err)
		return err
	}

	return nil
}
