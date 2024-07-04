// Package users manages users.
package users

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"cloud.google.com/go/firestore"

	"disruptive/config"
	"disruptive/lib/common"
)

// Document contains user fields.
type Document struct {
	Username    string         `firestore:"username" json:"username"`
	Domain      string         `firestore:"domain" json:"domain"`
	Preferences map[string]any `firestore:"preferences" json:"preferences,omitempty"`
	Tokens      map[string]int `firestore:"tokens" json:"tokens,omitempty"`
	CreatedAt   time.Time      `firestore:"created_at" json:"created_at"`
}

// Request contains a user.
type Request struct {
	Preferences map[string]any `json:"preferences"`
}

// FoldersRequest contains a user's folder tree.
type FoldersRequest map[string]any

// PreferencesRequest contains a user's preferences.
type PreferencesRequest map[string]any

func renderDocument(d Document) Document {
	t := map[string]int{}

	for k, v := range d.Tokens {
		t[strings.ReplaceAll(k, "--", ".")] = v
	}

	d.Tokens = t

	return d
}

// Add adds a user.
func Add(ctx context.Context, logCtx *slog.Logger, username string) error {
	logCtx = logCtx.With("fid", "sage.users.Add")

	client, err := firestore.NewClient(ctx, config.VARS.FirebaseProject)
	if err != nil {
		logCtx.Error("unable to create firestore client", "error", err)
		return err
	}
	defer client.Close()

	parts := strings.Split(username, "@")
	if len(parts) != 2 {
		return common.ErrBadRequest{Msg: "invalid username"}
	}

	d := Document{
		Username:  username,
		Domain:    parts[1],
		CreatedAt: time.Now(),
	}

	if _, err := client.Collection("sage").Doc(username).Create(ctx, d); err != nil {
		err = common.ConvertGRPCError(err)
		if errors.Is(err, common.ErrAlreadyExists{}) {
			logCtx.Warn("user already exists")
			return err
		}
		logCtx.Error("unable to create user document", "error", err)
		return err
	}

	return nil
}

// Get retrieves a user.
func Get(ctx context.Context, logCtx *slog.Logger, username string) (Document, error) {
	logCtx = logCtx.With("fid", "sage.users.Get")

	client, err := firestore.NewClient(ctx, config.VARS.FirebaseProject)
	if err != nil {
		logCtx.Error("unable to create firestore client", "error", err)
		return Document{}, err
	}
	defer client.Close()

	doc, err := client.Collection("sage").Doc(username).Get(ctx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to get user document", "error", err)
		return Document{}, err
	}

	d := Document{}
	if err := doc.DataTo(&d); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to read user data", "error", err)
		return Document{}, err
	}

	return renderDocument(d), nil
}

// Patch updates a user.
func Patch(ctx context.Context, logCtx *slog.Logger, username string, document Document) error {
	logCtx = logCtx.With("fid", "sage.users.Patch")

	client, err := firestore.NewClient(ctx, config.VARS.FirebaseProject)
	if err != nil {
		logCtx.Error("unable to create firestore client", "error", err)
		return err
	}
	defer client.Close()

	update := []firestore.Update{}

	if document.Preferences != nil {
		update = append(update, firestore.Update{Path: "preferences", Value: document.Preferences})
	}

	if len(update) < 1 {
		logCtx.Warn("nothing to update", "error", err)
		return nil
	}

	_, err = client.Collection("sage").Doc(username).Update(ctx, update)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to update user document", "error", err)
		return err
	}

	return nil
}
