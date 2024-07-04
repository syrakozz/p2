package users

import (
	"context"
	"log/slog"

	"disruptive/lib/users"
)

// Add creates a new user.
func Add(ctx context.Context, username, name, password string, permissions []string) error {
	logCtx := slog.With("username", username)

	u := users.User{
		Username:    username,
		Name:        name,
		Password:    password,
		Permissions: permissions,
	}

	return users.Add(ctx, logCtx, u)
}
