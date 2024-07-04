package users

import (
	"context"
	"log/slog"

	"disruptive/lib/users"
)

// Delete removes a user.
func Delete(ctx context.Context, username string) error {
	logCtx := slog.With("username", username)
	return users.Delete(ctx, logCtx, username)
}
