// Package projects manages projects.
package projects

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"

	"disruptive/config"
	"disruptive/lib/common"
)

// Document contains project fields.
type Document struct {
	ProjectID   string         `firestore:"project_id" json:"project_id"`
	Name        string         `firestore:"name" json:"name"`
	Status      string         `firestore:"status" json:"status"`
	Folders     map[string]any `firestore:"folders" json:"folders,omitempty"`
	Preferences map[string]any `firestore:"preferences" json:"preferences,omitempty"`
	Tokens      map[string]int `firestore:"tokens" json:"tokens,omitempty"`
	CreatedAt   time.Time      `firestore:"created_at" json:"created_at"`
}

// ProjectPreferencesReq contains project preferences.
type ProjectPreferencesReq map[string]any

// GetAll returns all projects for a user.
func GetAll(ctx context.Context, logCtx *slog.Logger, username, status string) ([]Document, error) {
	logCtx = logCtx.With("fid", "sage.projects.GetAll")

	client, err := firestore.NewClient(ctx, config.VARS.FirebaseProject)
	if err != nil {
		logCtx.Error("unable to create firestore client", "error", err)
		return nil, err
	}
	defer client.Close()

	path := fmt.Sprintf("sage/%s/projects", username)

	c := client.Collection(path)
	if c == nil {
		return nil, common.ErrNotFound{}
	}

	var q firestore.Query
	if status == "" {
		q = c.Where("status", "not-in", []string{"archived", "deleted"})
	} else {
		q = c.Where("status", "==", status)
	}

	docs, err := q.Documents(ctx).GetAll()
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to get user document", "error", err)
		return nil, err
	}

	documents := make([]Document, len(docs))

	for i := 0; i < len(docs); i++ {
		d := Document{}
		if err := docs[i].DataTo(&d); err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Error("unable to read user data", "error", err)
			return nil, err
		}

		clean := map[string]int{}
		for k, v := range d.Tokens {
			clean[strings.ReplaceAll(k, "--", ".")] = v
		}
		d.Tokens = clean

		documents[i] = d
	}

	return documents, nil
}

// Post creates a new project and returns the project ID.
func Post(ctx context.Context, logCtx *slog.Logger, username string, document Document) (string, error) {
	logCtx = logCtx.With("fid", "sage.projects.Post")

	client, err := firestore.NewClient(ctx, config.VARS.FirebaseProject)
	if err != nil {
		logCtx.Error("unable to create firestore client", "error", err)
		return "", err
	}
	defer client.Close()

	path := fmt.Sprintf("sage/%s/projects", username)

	c := client.Collection(path)
	if c == nil {
		logCtx.Error("collection", "error", "invalid path", "path", path)
		return "", common.ErrNotFound{}
	}

	document.ProjectID = uuid.New().String()
	document.CreatedAt = time.Now()

	if _, err := c.Doc(document.ProjectID).Set(ctx, document); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("set document", "error", err)
		return "", err
	}

	return document.ProjectID, nil
}

// Get retrieves a project
func Get(ctx context.Context, logCtx *slog.Logger, username, project string) (Document, error) {
	logCtx = logCtx.With("fid", "sage.projects.Get")

	client, err := firestore.NewClient(ctx, config.VARS.FirebaseProject)
	if err != nil {
		logCtx.Error("unable to create firestore client", "error", err)
		return Document{}, err
	}
	defer client.Close()

	path := fmt.Sprintf("sage/%s/projects", username)

	c := client.Collection(path)
	if c == nil {
		return Document{}, common.ErrNotFound{}
	}

	doc, err := c.Doc(project).Get(ctx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to get project document", "error", err)
		return Document{}, err
	}

	d := Document{}

	if err := doc.DataTo(&d); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to read project data", "error", err)
		return Document{}, err
	}

	clean := map[string]int{}
	for k, v := range d.Tokens {
		clean[strings.ReplaceAll(k, "--", ".")] = v
	}
	d.Tokens = clean

	return d, nil
}

// Patch updates a project.
func Patch(ctx context.Context, logCtx *slog.Logger, username, project string, document Document) error {
	logCtx = logCtx.With("fid", "sage.projects.Patch")

	client, err := firestore.NewClient(ctx, config.VARS.FirebaseProject)
	if err != nil {
		logCtx.Error("unable to create firestore client", "error", err)
		return err
	}
	defer client.Close()

	path := fmt.Sprintf("sage/%s/projects", username)

	c := client.Collection(path)
	if c == nil {
		return common.ErrNotFound{}
	}

	update := []firestore.Update{}

	if document.Name != "" {
		update = append(update, firestore.Update{Path: "name", Value: document.Name})
	}

	if document.Status != "" {
		update = append(update, firestore.Update{Path: "status", Value: document.Status})
	}

	if document.Folders != nil {
		update = append(update, firestore.Update{Path: "folders", Value: document.Folders})
	}

	if document.Preferences != nil {
		update = append(update, firestore.Update{Path: "preferences", Value: document.Preferences})
	}

	if len(update) < 1 {
		logCtx.Warn("nothing to update", "error", err)
		return nil
	}

	_, err = c.Doc(project).Update(ctx, update)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to update project document", "error", err)
		return err
	}

	return nil
}
