// Package firebase ...
package firebase

import (
	"context"
	"disruptive/config"
	"log/slog"

	fb "firebase.google.com/go"
	"firebase.google.com/go/auth"
)

var (
	// App is the Firebase app context.
	App        *fb.App
	authClient *auth.Client

	// GCSBucket is shared bucket for firebase commands.
	GCSBucket = config.VARS.FirebaseProject + ".appspot.com"
)

func init() {
	var err error

	App, err = fb.NewApp(context.Background(), nil)
	if App == nil || err != nil {
		slog.Warn("unable to create firebase app", "error", err)
		return
	}

	authClient, err = App.Auth(context.Background())

	if authClient == nil || err != nil {
		slog.Warn("unable to create firebase auth", "error", err)
		return
	}
}
