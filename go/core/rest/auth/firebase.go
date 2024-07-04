package auth

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"firebase.google.com/go/auth"

	"disruptive/lib/common"
	"disruptive/lib/firebase"
)

// GetFirebaseUser verifies the Authorizatoin Bearer JWT token and returns the user.
//
// It will also fallback and return a user if the X-UID header is set.
// X-UID should only be used when developing or testing and never used by the UI.
func GetFirebaseUser(ctx context.Context, logCtx *slog.Logger, header http.Header) (*auth.UserRecord, error) {
	fid := slog.String("fid", "rest.auth.GetFirebaseUser")

	bearer := strings.Split(header.Get("Authorization"), "Bearer ")
	if len(bearer) == 2 {
		user, err := firebase.GetUserByJWT(ctx, logCtx, bearer[1])
		if err != nil {
			logCtx.Error("unable to get user by jwt", fid, "error", err)
			return nil, err
		}

		return user, nil
	}

	uid := header.Get("X-UID")
	if uid != "" {
		user, err := firebase.GetUserByUID(ctx, logCtx, uid)
		if err != nil {
			logCtx.Error("unable to get user by uid", fid, "error", err)
			return nil, err
		}

		return user, nil
	}

	logCtx.Error("unauthorized", fid)
	return nil, common.ErrUnauthorized
}
