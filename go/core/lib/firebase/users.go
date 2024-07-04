package firebase

import (
	"context"
	"log/slog"

	"firebase.google.com/go/auth"
)

// GetUserByJWT verifies a JWT Firebase token and returns the user.
func GetUserByJWT(ctx context.Context, logCtx *slog.Logger, jwt string) (*auth.UserRecord, error) {
	logCtx = logCtx.With("fid", "firebase.GetUserByJWT")

	token, err := authClient.VerifyIDToken(ctx, jwt)
	if err != nil {
		logCtx.Error("unable to verify jwt", "error", err)
		return nil, err
	}

	return GetUserByUID(ctx, logCtx, token.UID)
}

// GetUserByUID returns a Firebase user given a UID.
func GetUserByUID(ctx context.Context, logCtx *slog.Logger, uid string) (*auth.UserRecord, error) {
	logCtx = logCtx.With("fid", "firebase.GetUserByUID")

	user, err := authClient.GetUser(ctx, uid)
	if err != nil {
		logCtx.Error("unable to get user", "error", err)
		return nil, err
	}

	return user, nil
}

// SetUserClaimsByUID replaces the custom claims for a user by UID and returns the modified user.
func SetUserClaimsByUID(ctx context.Context, logCtx *slog.Logger, uid string, claims map[string]any) (*auth.UserRecord, error) {
	logCtx = logCtx.With("fid", "firebase.SetUserClaimsByUID")

	if err := authClient.SetCustomUserClaims(ctx, uid, claims); err != nil {
		logCtx.Error("unable to set user claims")
		return nil, err
	}

	return GetUserByUID(ctx, logCtx, uid)
}

// AddUserClaimsByUID modifies specific custom claims for a user by UID and returns the modified user.
func AddUserClaimsByUID(ctx context.Context, logCtx *slog.Logger, uid string, claims map[string]any) (*auth.UserRecord, error) {
	logCtx = logCtx.With("fid", "firebase.AddUserClaimsByUID")

	user, err := GetUserByUID(ctx, logCtx, uid)
	if err != nil {
		return nil, err
	}

	if user.CustomClaims == nil {
		user.CustomClaims = claims
	} else {
		for k, v := range claims {
			user.CustomClaims[k] = v
		}
	}

	if err := authClient.SetCustomUserClaims(ctx, uid, user.CustomClaims); err != nil {
		logCtx.Error("unable to set user claims", "error", err)
		return nil, err
	}

	return GetUserByUID(ctx, logCtx, uid)
}

// DeleteAllUserClaimsByUID deletes all custom claims for a user by UID and returns the modified user.
func DeleteAllUserClaimsByUID(ctx context.Context, logCtx *slog.Logger, uid string) (*auth.UserRecord, error) {
	logCtx = logCtx.With("fid", "firebase.DeleteAllUserClaimsByUID")

	if err := authClient.SetCustomUserClaims(ctx, uid, nil); err != nil {
		logCtx.Error("unable to set claims", "error", err)
		return nil, err
	}

	return GetUserByUID(ctx, logCtx, uid)
}

// DeleteUserClaimsByUID deletes specific custom claims for a user by UID and returns the modified user.
func DeleteUserClaimsByUID(ctx context.Context, logCtx *slog.Logger, uid string, claims []string) (*auth.UserRecord, error) {
	logCtx = logCtx.With("fid", "firebase.DeleteUserClaimsByUID")

	user, err := GetUserByUID(ctx, logCtx, uid)
	if err != nil {
		logCtx.Error("unable to get user", "error", err)
		return nil, err
	}

	if user.CustomClaims == nil {
		return user, nil
	}

	for _, k := range claims {
		delete(user.CustomClaims, k)
	}

	if err := authClient.SetCustomUserClaims(ctx, uid, user.CustomClaims); err != nil {
		logCtx.Error("unable to set claims", "error", err)
		return nil, err
	}

	return GetUserByUID(ctx, logCtx, uid)
}

// DeleteUser delete a specific user based on UID.
func DeleteUser(ctx context.Context, uid string) error {
	return authClient.DeleteUser(ctx, uid)
}
