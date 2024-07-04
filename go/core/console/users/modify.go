package users

import (
	"context"
	"log/slog"

	"disruptive/lib/users"
)

// Modify modifies a user.
func Modify(ctx context.Context, username, name, password string, permissions []string) error {
	logCtx := slog.With("username", username)

	u := users.User{
		Username:    username,
		Name:        name,
		Password:    password,
		Permissions: permissions,
	}

	return users.Modify(ctx, logCtx, u)
}
