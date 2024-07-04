// Package firestore integrates with the firestore database.
package firestore

import (
	"context"
	"log/slog"

	"cloud.google.com/go/firestore"

	"disruptive/config"
)

var (
	// Client is the shared firestore client.
	Client *firestore.Client
)

func init() {
	var err error

	Client, err = firestore.NewClient(context.Background(), config.VARS.FirebaseProject)
	if Client == nil || err != nil {
		slog.Warn("unable to create firestore client", "error", err)
	}
}
