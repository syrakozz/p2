package models

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"

	"disruptive/config"
	"disruptive/lib/common"
	"disruptive/lib/configs"
	"disruptive/lib/openai"
	"disruptive/lib/sage/models/chat"
	"disruptive/lib/sage/sessions"
)

// Document contains model fields.
type Document struct {
	ProjectID      string         `firestore:"project_id" json:"project_id"`
	SessionID      string         `firestore:"session_id" json:"session_id"`
	ModelID        string         `firestore:"model_id" json:"model_id"`
	Name           string         `firestore:"name" json:"name"`
	Model          string         `firestore:"model" json:"model"`
	Type           string         `firestore:"type" json:"type"`
	Creativity     int            `firestore:"creativity" json:"creativity"`
	Personality    string         `firestore:"personality" json:"personality"`
	PersonalityKey string         `firestore:"personality_key" json:"personality_key"`
	MaxTokens      int            `firestore:"max_tokens" json:"max_tokens"`
	Status         string         `firestore:"status" json:"status"`
	LastEntry      int            `firestore:"last_entry" json:"last_entry"`
	LastBlock      int            `firestore:"last_block" json:"last_block"`
	LastSummary    int            `firestore:"last_summary" json:"last_summary"`
	Preferences    map[string]any `firestore:"preferences,omitempty" json:"preferences,omitempty"`
	CreatedAt      time.Time      `firestore:"created_at" json:"created_at"`
}

// Request contains model info.
type Request struct {
	Name        string `json:"name"`
	Model       string `json:"model"`
	Type        string `json:"type"`
	Creativity  int    `json:"creativity"`
	Personality string `json:"personality"`
	MaxTokens   int    `json:"max_tokens"`
}

// DocumentMap contains model info in map instead of a struct.
type DocumentMap map[string]any

// EntryMap contains entry info in a map instead of a struct.
type EntryMap map[string]any

// EntryVersionResponse contains version response information.
type EntryVersionResponse struct {
	Entry                 string `json:"entry"`
	Version               string `json:"version"`
	Text                  string `json:"text,omitempty"`
	FinishReason          string `json:"finish_reason,omitempty"`
	TokensVersionPrompt   int    `json:"tokens_version_prompt"`
	TokensVersionResponse int    `json:"tokens_version_response"`
	TokensEntryPrompts    int    `json:"tokens_entry_prompts"`
	TokensEntryResponses  int    `json:"tokens_entry_responses"`
	TokensModelPrompts    int    `json:"tokens_model_prompts"`
	TokensModelResponses  int    `json:"tokens_model_responses"`
}

// Post creates a new model and returns the model ID.
func Post(ctx context.Context, logCtx *slog.Logger, username, project, session string, document Document) (string, error) {
	logCtx = logCtx.With("fid", "sage.models.Post")

	client, err := firestore.NewClient(ctx, config.VARS.FirebaseProject)
	if err != nil {
		logCtx.Error("unable to create firestore client", "error", err)
		return "", err
	}
	defer client.Close()

	path := fmt.Sprintf("sage/%s/projects/%s/sessions/%s/models", username, project, session)

	c := client.Collection(path)
	if c == nil {
		logCtx.Error("collection", "error", err, "path", path)
		return "", common.ErrNotFound{}
	}

	document.ProjectID = project
	document.SessionID = session
	document.ModelID = uuid.New().String()
	document.Status = "in progress"
	document.LastEntry = -1
	document.LastBlock = -1
	document.LastSummary = -1
	document.CreatedAt = time.Now()

	if document.PersonalityKey != "" {
		c, err := configs.Get(ctx, logCtx, "personalities")
		if err != nil {
			return "", common.ErrConsistency
		}
		p, ok := c[document.PersonalityKey].(string)
		if !ok {
			return "", common.ErrConsistency
		}
		document.Personality = p
	}

	if _, err := c.Doc(document.ModelID).Set(ctx, document); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("set document", "error", err)
		return "", err
	}

	// Update session

	s := sessions.Document{
		Models: sessions.Models{
			document.ModelID: {
				Name:      document.Name,
				Model:     document.Model,
				Type:      document.Type,
				Status:    document.Status,
				CreatedAt: document.CreatedAt,
			},
		},
	}

	if err := sessions.Patch(ctx, logCtx, username, project, session, s); err != nil {
		logCtx.Error("unable to set model in session", "error", err)
		return "", common.ErrNotFound{}
	}

	return document.ModelID, nil
}

// Get retrieves a model.
func Get(ctx context.Context, logCtx *slog.Logger, username, project, session, model string) (DocumentMap, error) {
	logCtx = logCtx.With("fid", "sage.models.Get")

	client, err := firestore.NewClient(ctx, config.VARS.FirebaseProject)
	if err != nil {
		logCtx.Error("unable to create firestore client", "error", err)
		return nil, err
	}
	defer client.Close()

	path := fmt.Sprintf("sage/%s/projects/%s/sessions/%s/models", username, project, session)

	c := client.Collection(path)
	if c == nil {
		return nil, common.ErrNotFound{}
	}

	doc, err := c.Doc(model).Get(ctx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to get model document", "error", err)
		return nil, err
	}

	d := DocumentMap{}

	if err := doc.DataTo(&d); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to read model data", "error", err)
		return nil, err
	}

	return d, nil
}

// Patch updates a model.
func Patch(ctx context.Context, logCtx *slog.Logger, username, project, session, model string, document Document) error {
	logCtx = logCtx.With("fid", "sage.models.Patch")

	client, err := firestore.NewClient(ctx, config.VARS.FirebaseProject)
	if err != nil {
		logCtx.Error("unable to create firestore client", "error", err)
		return err
	}
	defer client.Close()

	path := fmt.Sprintf("sage/%s/projects/%s/sessions/%s/models", username, project, session)

	c := client.Collection(path)
	if c == nil {
		return common.ErrNotFound{}
	}

	updates := []firestore.Update{}

	if document.Name != "" {
		updates = append(updates, firestore.Update{Path: "name", Value: document.Name})
	}

	if document.Creativity != 0 {
		if document.Creativity < 0 {
			document.Creativity = 0
		}
		updates = append(updates, firestore.Update{Path: "creativity", Value: document.Creativity})
	}

	if document.PersonalityKey != "" {
		c, err := configs.Get(ctx, logCtx, "personalities")
		if err != nil {
			return common.ErrConsistency
		}
		p, ok := c[document.PersonalityKey].(string)
		if !ok {
			return common.ErrConsistency
		}
		updates = append(updates, firestore.Update{Path: "personality", Value: p})
		updates = append(updates, firestore.Update{Path: "personality_key", Value: document.PersonalityKey})
	} else if document.Personality != "" {
		updates = append(updates, firestore.Update{Path: "personality", Value: document.Personality})
		updates = append(updates, firestore.Update{Path: "personality_key", Value: ""})
	}

	if document.Status != "" {
		updates = append(updates, firestore.Update{Path: "status", Value: document.Status})
	}

	if document.Preferences != nil {
		updates = append(updates, firestore.Update{Path: "preferences", Value: document.Preferences})
	}

	if len(updates) < 1 {
		logCtx.Warn("nothing to update", "error", err)
		return nil
	}

	_, err = c.Doc(model).Update(ctx, updates)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to update session document", "error", err)
		return err
	}

	// Update session.

	if document.Name != "" || document.Status != "" {
		path = fmt.Sprintf("sage/%s/projects/%s/sessions", username, project)

		c = client.Collection(path)
		if c == nil {
			return common.ErrNotFound{}
		}

		updates = []firestore.Update{}

		if document.Name != "" {
			updates = append(updates, firestore.Update{Path: fmt.Sprintf("models.%s.name", model), Value: document.Name})
		}

		if document.Status != "" {
			updates = append(updates, firestore.Update{Path: fmt.Sprintf("models.%s.status", model), Value: document.Status})
		}

		if len(updates) > 0 {
			if _, err = c.Doc(session).Update(ctx, updates); err != nil {
				err = common.ConvertGRPCError(err)
				logCtx.Warn("unable to update session document", "error", err)
				return err
			}
		}
	}

	return nil
}

// PostEntry creates a new version.
func PostEntry(ctx context.Context, logCtx *slog.Logger, username, project, session, model, entry, content string) (EntryVersionResponse, error) {
	logCtx = logCtx.With("fid", "sage.models.PostEntry")

	client, err := firestore.NewClient(ctx, config.VARS.FirebaseProject)
	if err != nil {
		logCtx.Error("unable to create firestore client", "error", err)
		return EntryVersionResponse{}, err
	}
	defer client.Close()

	// Update model

	m, err := chat.GetGPT(ctx, logCtx, client, username, project, session, model)
	if err != nil {
		logCtx.Error("unable to get model", "error", err)
		return EntryVersionResponse{}, err
	}

	path := fmt.Sprintf("sage/%s/projects/%s/sessions/%s/models", username, project, session)

	c := client.Collection(path)
	if c == nil {
		return EntryVersionResponse{}, common.ErrNotFound{}
	}

	v := m.Entries[entry].LastVersion + 1

	updates := []firestore.Update{
		{Path: fmt.Sprintf("entries.%s.current_version", entry), Value: v},
		{Path: fmt.Sprintf("entries.%s.last_version", entry), Value: v},
		{Path: fmt.Sprintf("entries.%s.versions.%02d.content", entry, v), Value: content},
		{Path: fmt.Sprintf("entries.%s.versions.%02d.timestamp", entry, v), Value: time.Now()},
	}

	if _, err = c.Doc(model).Update(ctx, updates); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to update model entry version", "error", err)
		return EntryVersionResponse{}, err
	}

	res := EntryVersionResponse{
		Entry:   entry,
		Version: fmt.Sprintf("%02d", v),
	}

	return res, nil
}

// GetEntry retrieves a model.
func GetEntry(ctx context.Context, logCtx *slog.Logger, username, project, session, model, entry string) (EntryMap, error) {
	logCtx = logCtx.With("fid", "sage.models.GetEntry")

	client, err := firestore.NewClient(ctx, config.VARS.FirebaseProject)
	if err != nil {
		logCtx.Error("unable to create firestore client", "error", err)
		return nil, err
	}
	defer client.Close()

	path := fmt.Sprintf("sage/%s/projects/%s/sessions/%s/models", username, project, session)

	c := client.Collection(path)
	if c == nil {
		return nil, common.ErrNotFound{}
	}

	doc, err := c.Doc(model).Get(ctx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to get model document", "error", err)
		return nil, err
	}

	d := DocumentMap{}

	if err := doc.DataTo(&d); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to read model data", "error", err)
		return nil, err
	}

	// Copy parts of ModelRes to em
	em := EntryMap{}

	ee, ok := d["entries"].(EntryMap)
	if !ok {
		return nil, common.ErrNotFound{}
	}

	e, ok := ee[entry]
	if !ok {
		return nil, common.ErrNotFound{}
	}

	em["model_id"] = d["model_id"]
	em["entry_id"] = entry
	em["entry"] = e
	em["tokens_prompts"] = d["tokens_prompts"]
	em["tokens_responses"] = d["tokens_responses"]

	return em, nil
}

// PostEntryVersion creates a new entry version based on an existing entry version.
func PostEntryVersion(ctx context.Context, logCtx *slog.Logger, username, project, session, model, entry, version, words, bullets string) (EntryVersionResponse, error) {
	logCtx = logCtx.With("fid", "sage.models.PostEntryVersion")

	client, err := firestore.NewClient(ctx, config.VARS.FirebaseProject)
	if err != nil {
		logCtx.Error("unable to create firestore client", "error", err)
		return EntryVersionResponse{}, err
	}
	defer client.Close()

	m, err := chat.GetGPT(ctx, logCtx, client, username, project, session, model)
	if err != nil {
		logCtx.Error("unable to get model", "error", err)
		return EntryVersionResponse{}, err
	}

	e, ok := m.Entries[entry]
	if !ok {
		return EntryVersionResponse{}, common.ErrBadRequest{Msg: "invalid entry"}
	}

	v, ok := e.Versions[version]
	if !ok {
		return EntryVersionResponse{}, common.ErrBadRequest{Msg: "invalid version"}
	}

	newV := e.LastVersion + 1

	var prompt string

	switch {
	case words != "":
		prompt = fmt.Sprintf("Create a %s-word summary of the following.\n\n", words) + v.Content
	case bullets != "":
		prompt = fmt.Sprintf("Create a %s-bullet list summary of the following.\n\n", bullets) + v.Content
	default:
		return EntryVersionResponse{}, common.ErrBadRequest{Msg: "invalid entry"}
	}

	chatReq := openai.ChatRequest{
		Model:    "gpt-3.5-turbo",
		Messages: []openai.ChatMessage{{Role: "user", Content: prompt}},
	}

	chatRes, err := openai.PostChat(ctx, logCtx, chatReq)
	if err != nil {
		logCtx.Error("unable to post openai chat.")
		return EntryVersionResponse{}, err
	}

	path := fmt.Sprintf("sage/%s/projects/%s/sessions/%s/models", username, project, session)

	c := client.Collection(path)
	if c == nil {
		return EntryVersionResponse{}, common.ErrNotFound{}
	}

	// Update model

	updates := []firestore.Update{
		{Path: "tokens_prompts", Value: firestore.Increment(chatRes.UsagePrompt)},
		{Path: "tokens_responses", Value: firestore.Increment(chatRes.UsageResponse)},
		{Path: fmt.Sprintf("entries.%s.current_version", entry), Value: newV},
		{Path: fmt.Sprintf("entries.%s.last_version", entry), Value: newV},
		{Path: fmt.Sprintf("entries.%s.tokens_prompts", entry), Value: firestore.Increment(chatRes.UsagePrompt)},
		{Path: fmt.Sprintf("entries.%s.tokens_responses", entry), Value: firestore.Increment(chatRes.UsageResponse)},
		{Path: fmt.Sprintf("entries.%s.versions.%02d.content", entry, newV), Value: chatRes.Text},
		{Path: fmt.Sprintf("entries.%s.versions.%02d.timestamp", entry, newV), Value: time.Now()},
		{Path: fmt.Sprintf("entries.%s.versions.%02d.tokens_prompt", entry, newV), Value: chatRes.UsagePrompt},
		{Path: fmt.Sprintf("entries.%s.versions.%02d.tokens_response", entry, newV), Value: chatRes.UsageResponse},
		{Path: fmt.Sprintf("entries.%s.versions.%02d.summary", entry, newV), Value: true},
	}

	if _, err = c.Doc(model).Update(ctx, updates); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to update model entry version", "error", err)
		return EntryVersionResponse{}, err
	}

	// Update session

	path = fmt.Sprintf("sage/%s/projects/%s/sessions", username, project)

	c = client.Collection(path)
	if c == nil {
		return EntryVersionResponse{}, common.ErrNotFound{}
	}

	updates = []firestore.Update{
		{Path: fmt.Sprintf("models.%s.tokens_prompts", model), Value: firestore.Increment(chatRes.UsagePrompt)},
		{Path: fmt.Sprintf("models.%s.tokens_responses", model), Value: firestore.Increment(chatRes.UsageResponse)},
	}

	if _, err = c.Doc(session).Update(ctx, updates); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to update session document", "error", err)
		return EntryVersionResponse{}, err
	}

	// Update project

	path = fmt.Sprintf("sage/%s/projects", username)

	c = client.Collection(path)
	if c == nil {
		logCtx.Error("collection not found", "path", path)
		return EntryVersionResponse{}, common.ErrNotFound{}
	}

	modelEncoded := strings.ReplaceAll(m.Model, ".", "--")

	updates = []firestore.Update{
		{Path: fmt.Sprintf("tokens.%s_prompts", modelEncoded), Value: firestore.Increment(chatRes.UsagePrompt)},
		{Path: fmt.Sprintf("tokens.%s_responses", modelEncoded), Value: firestore.Increment(chatRes.UsageResponse)},
	}

	if _, err = c.Doc(project).Update(ctx, updates); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to update project document", "error", err)
		return EntryVersionResponse{}, err
	}

	// Update user

	c = client.Collection("sage")
	if c == nil {
		logCtx.Error("collection not found", "path", "sage")
		return EntryVersionResponse{}, common.ErrNotFound{}
	}
	if _, err = c.Doc(username).Update(ctx, updates); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to update user document", "error", err)
		return EntryVersionResponse{}, err
	}

	res := EntryVersionResponse{
		Entry:                 entry,
		Version:               fmt.Sprintf("%02d", newV),
		Text:                  chatRes.Text,
		FinishReason:          chatRes.FinishReason,
		TokensVersionPrompt:   chatRes.UsagePrompt,
		TokensVersionResponse: chatRes.UsageResponse,
		TokensEntryPrompts:    chatRes.UsagePrompt + e.TokensPrompts,
		TokensEntryResponses:  chatRes.UsageResponse + e.TokensResponses,
		TokensModelPrompts:    chatRes.UsagePrompt + m.TokensPrompts,
		TokensModelResponses:  chatRes.UsageResponse + m.TokensResponses,
	}

	return res, nil
}

// PutEntryVersion changes the current entry version.
func PutEntryVersion(ctx context.Context, logCtx *slog.Logger, username, project, session, model, entry, version string) (EntryVersionResponse, error) {
	logCtx = logCtx.With("fid", "sage.models.PutEntryVersion")

	client, err := firestore.NewClient(ctx, config.VARS.FirebaseProject)
	if err != nil {
		logCtx.Error("unable to create firestore client", "error", err)
		return EntryVersionResponse{}, err
	}
	defer client.Close()

	// Update model

	m, err := chat.GetGPT(ctx, logCtx, client, username, project, session, model)
	if err != nil {
		logCtx.Error("unable to get model", "error", err)
		return EntryVersionResponse{}, err
	}

	path := fmt.Sprintf("sage/%s/projects/%s/sessions/%s/models", username, project, session)

	c := client.Collection(path)
	if c == nil {
		return EntryVersionResponse{}, common.ErrNotFound{}
	}

	e, ok := m.Entries[entry]
	if !ok {
		return EntryVersionResponse{}, common.ErrBadRequest{Msg: "invalid entry"}
	}

	v, err := strconv.Atoi(version)
	if err != nil {
		return EntryVersionResponse{}, common.ErrBadRequest{Msg: "invalid version"}
	}

	if e.LastVersion < 0 || v < 0 || v > e.LastVersion {
		return EntryVersionResponse{}, common.ErrBadRequest{Msg: "invalid version"}
	}

	updates := []firestore.Update{
		{Path: fmt.Sprintf("entries.%s.current_version", entry), Value: v},
	}

	if _, err = c.Doc(model).Update(ctx, updates); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to update model entry current version", "error", err)
		return EntryVersionResponse{}, err
	}

	res := EntryVersionResponse{
		Entry:   entry,
		Version: fmt.Sprintf("%02d", v),
	}

	return res, nil
}
