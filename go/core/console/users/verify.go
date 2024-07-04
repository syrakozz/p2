package users

import (
	"context"
	"fmt"
	"log/slog"

	"disruptive/lib/users"
)

// VerifyPassword validates a user's password.
func VerifyPassword(ctx context.Context, username, password string) error {
	logCtx := slog.With("username", username)

	ok, err := users.VerifyPassword(ctx, logCtx, username, password)
	if err != nil || !ok {
		fmt.Println("Not verified")
	} else {
		fmt.Println("Verified")
	}

	return nil
}
