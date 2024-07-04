// Package sessions manages sessions.
package sessions

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"

	"disruptive/config"
	"disruptive/lib/common"
)

// Models contains a summary of associated models.
type Models map[string]struct {
	Name            string    `firestore:"name" json:"name"`
	Model           string    `firestore:"model" json:"model"`
	Type            string    `firestore:"type" json:"type"`
	Status          string    `firestore:"status" json:"status"`
	TokensPrompts   int       `firestore:"tokens_prompts" json:"token_prompts"`
	TokensResponses int       `firestore:"tokens_responses" json:"token_reponses"`
	CreatedAt       time.Time `firestore:"created_at" json:"created_at"`
}

// Document contains session fields.
type Document struct {
	ProjectID   string         `firestore:"project_id" json:"project_id"`
	SessionID   string         `firestore:"session_id" json:"session_id"`
	Name        string         `firestore:"name" json:"name"`
	Status      string         `firestore:"status" json:"status"`
	Folder      []string       `firestore:"folder,omitempty" json:"folder,omitempty"`
	Preferences map[string]any `firestore:"preferences,omitempty" json:"preferences,omitempty"`
	Models      Models         `firestore:"models,omitempty" json:"models,omitempty"`
	CreatedAt   time.Time      `firestore:"created_at" json:"created_at"`
}

// GetAll returns all sessions for a project.
func GetAll(ctx context.Context, logCtx *slog.Logger, username, project, status string) ([]Document, error) {
	logCtx = logCtx.With("fid", "sage.sessions.GetAll")

	client, err := firestore.NewClient(ctx, config.VARS.FirebaseProject)
	if err != nil {
		logCtx.Error("unable to create firestore client", "error", err)
		return nil, err
	}
	defer client.Close()

	path := fmt.Sprintf("sage/%s/projects/%s/sessions", username, project)

	c := client.Collection(path)
	if c == nil {
		logCtx.Error("collection", "error", "invalid path", "path", path)
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
		documents[i] = d
	}

	return documents, nil
}

// Post creates a new session and returns the session ID.
func Post(ctx context.Context, logCtx *slog.Logger, username, project string, document Document) (string, error) {
	logCtx = logCtx.With("fid", "sage.sessions.Post")

	client, err := firestore.NewClient(ctx, config.VARS.FirebaseProject)
	if err != nil {
		logCtx.Error("unable to create firestore client", "error", err)
		return "", err
	}
	defer client.Close()

	path := fmt.Sprintf("sage/%s/projects/%s/sessions", username, project)

	c := client.Collection(path)
	if c == nil {
		logCtx.Error("collection", "error", "invalid path", "path", path)
		return "", common.ErrNotFound{}
	}

	document.ProjectID = project
	document.SessionID = uuid.New().String()
	document.CreatedAt = time.Now()

	if _, err := c.Doc(document.SessionID).Set(ctx, document); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("set document", "error", err)
		return "", err
	}

	return document.SessionID, nil
}

// Get retrieves a session
func Get(ctx context.Context, logCtx *slog.Logger, username, project, session string) (Document, error) {
	logCtx = logCtx.With("fid", "sage.sessions.Get")

	client, err := firestore.NewClient(ctx, config.VARS.FirebaseProject)
	if err != nil {
		logCtx.Error("unable to create firestore client", "error", err)
		return Document{}, err
	}
	defer client.Close()

	path := fmt.Sprintf("sage/%s/projects/%s/sessions", username, project)

	c := client.Collection(path)
	if c == nil {
		return Document{}, common.ErrNotFound{}
	}

	doc, err := c.Doc(session).Get(ctx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to get session document", "error", err)
		return Document{}, err
	}

	d := Document{}

	if err := doc.DataTo(&d); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to read session data", "error", err)
		return Document{}, err
	}

	return d, nil
}

// Patch updates a session.
func Patch(ctx context.Context, logCtx *slog.Logger, username, project, session string, document Document) error {
	logCtx = logCtx.With("fid", "sage.sessions.Patch")

	client, err := firestore.NewClient(ctx, config.VARS.FirebaseProject)
	if err != nil {
		logCtx.Error("unable to create firestore client", "error", err)
		return err
	}
	defer client.Close()

	path := fmt.Sprintf("sage/%s/projects/%s/sessions", username, project)

	c := client.Collection(path)
	if c == nil {
		return common.ErrNotFound{}
	}

	updates := []firestore.Update{}

	if document.Name != "" {
		updates = append(updates, firestore.Update{Path: "name", Value: document.Name})
	}

	if document.Status != "" {
		updates = append(updates, firestore.Update{Path: "status", Value: document.Status})
	}

	if document.Folder != nil {
		updates = append(updates, firestore.Update{Path: "folder", Value: document.Folder})
	}

	if document.Preferences != nil {
		updates = append(updates, firestore.Update{Path: "preferences", Value: document.Preferences})
	}

	if document.Models != nil {
		for u, m := range document.Models {
			updates = append(updates, firestore.Update{Path: "models." + u, Value: m})
		}
	}

	if len(updates) < 1 {
		logCtx.Warn("nothing to update", "error", err)
		return nil
	}

	_, err = c.Doc(session).Update(ctx, updates)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to update session document", "error", err)
		return err
	}

	return nil
}
