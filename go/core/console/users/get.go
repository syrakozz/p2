package users

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"disruptive/lib/common"
	"disruptive/lib/users"
)

// Get retrieves a user.
func Get(ctx context.Context, username string) error {
	logCtx := slog.With("username", username)

	user, err := users.Get(ctx, logCtx, username)
	if err != nil {
		if errors.Is(err, common.ErrNotFound{}) {
			logCtx.Info("username not found")
			return nil
		}

		logCtx.Error("unable to get user", "error", err)
		return err
	}

	fmt.Println("Username:", username)
	fmt.Println("Name:", user.Name)
	fmt.Println("Permissions:", strings.Join(user.Permissions, ", "))
	return nil
}
