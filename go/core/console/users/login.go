package users

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"

	"disruptive/config"
	"disruptive/lib/users"
)

// Login creates a new user.
func Login(ctx context.Context, username string, expire time.Duration) error {
	logCtx := slog.With("username", username)

	loginSession := uuid.New().String()
	if err := users.SetLoginSession(ctx, logCtx, username, loginSession); err != nil {
		logCtx.Error("unable to set login session", "error", err)
		return err
	}

	u, err := users.Get(ctx, logCtx, username)
	if err != nil {
		logCtx.Error("unable to get user", "error", err)
		return err
	}

	if expire.Seconds() == 0.0 {
		expire, _ = time.ParseDuration("876000h") // 100 years
	}
	exp := time.Now().Add(expire)
	logCtx = logCtx.With("expire", exp.Format(time.DateTime))

	claims := &users.JWTClaims{
		Username:     username,
		Name:         u.Name,
		Permissions:  u.Permissions,
		LoginSession: loginSession,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(config.VARS.JWTSessionSecret))
	if err != nil {
		logCtx.Error("unable to create login token", "error", err)
		return err
	}

	logCtx.Info("login")
	fmt.Println()
	fmt.Println(t)
	return nil
}
