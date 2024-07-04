package chatpad

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"cloud.google.com/go/firestore"

	"disruptive/config"
	"disruptive/lib/common"
	"disruptive/lib/openai"
)

// GPTEntryVersion contains a single entry version in the database.
type GPTEntryVersion struct {
	Content        string    `firestore:"content"`
	Timestamp      time.Time `firestore:"timestamp"`
	TokensPrompt   int       `firestore:"tokens_prompt"`
	TokensResponse int       `firestore:"tokens_response"`
	Personality    string    `firestore:"personality"`
	Prompts        []string  `firestore:"prompts"`
}

// GPTEntry contains a single entry in the database.
type GPTEntry struct {
	Role            string                     `firestore:"role"`
	LastVersion     int                        `firestore:"last_version"`
	CurrentVersion  int                        `firestore:"current_version"`
	TokensPrompts   int                        `firestore:"tokens_prompts"`
	TokensResponses int                        `firestore:"tokens_responses"`
	Enabled         bool                       `firestore:"enabled"`
	Versions        map[string]GPTEntryVersion `firestore:"versions"`
}

// GPTDocument contains the full response from the database.
type GPTDocument struct {
	ProjectID       string              `firestore:"project_id"`
	SessionID       string              `firestore:"session_id"`
	ModelID         string              `firestore:"model_id"`
	Name            string              `firestore:"name"`
	Model           string              `firestore:"model"`
	Type            string              `firestore:"type"`
	Creativity      int                 `firestore:"creativity"`
	Personality     string              `firestore:"personality"`
	MaxTokens       int                 `firestore:"max_tokens"`
	Status          string              `firestore:"status"`
	LastEntry       int                 `firestore:"last_entry"`
	TokensPrompts   int                 `firestore:"tokens_prompts"`
	TokensResponses int                 `firestore:"tokens_responses"`
	Entries         map[string]GPTEntry `firestore:"entries"`
}

// GetGPT retrieves a ChatGPT model.
func GetGPT(ctx context.Context, logCtx *slog.Logger, client *firestore.Client, username, project, session, model string) (GPTDocument, error) {
	logCtx = logCtx.With("fid", "sage.models.chatpad.GetGPT")

	path := fmt.Sprintf("sage/%s/projects/%s/sessions/%s/models", username, project, session)

	c := client.Collection(path)
	if c == nil {
		return GPTDocument{}, common.ErrNotFound{}
	}

	doc, err := c.Doc(model).Get(ctx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to get model document", "error", err)
		return GPTDocument{}, err
	}

	m := GPTDocument{}

	if err := doc.DataTo(&m); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to read model data", "error", err)
		return GPTDocument{}, err
	}

	return m, nil
}

// AddGPTAssistantEntry adds a new assistant entry.
func AddGPTAssistantEntry(ctx context.Context, logCtx *slog.Logger,
	username, project, session, model, message string, tokensPrompt, tokensResponse int) error {
	logCtx = logCtx.With("fid", "sage.models.chatpad.AddGPTAssistantEntry")

	client, err := firestore.NewClient(ctx, config.VARS.FirebaseProject)
	if err != nil {
		logCtx.Error("unable to create firestore client", "error", err)
		return err
	}
	defer client.Close()

	// Update model

	m, err := GetGPT(ctx, logCtx, client, username, project, session, model)
	if err != nil {
		logCtx.Error("unable to get model", "error", err)
		return err
	}

	path := fmt.Sprintf("sage/%s/projects/%s/sessions/%s/models", username, project, session)

	c := client.Collection(path)
	if c == nil {
		return common.ErrNotFound{}
	}

	entry := m.LastEntry + 1

	updates := []firestore.Update{
		{Path: "last_entry", Value: entry},
		{Path: "tokens_prompts", Value: firestore.Increment(tokensPrompt)},
		{Path: "tokens_responses", Value: firestore.Increment(tokensResponse)},
		{Path: fmt.Sprintf("entries.%04d.role", entry), Value: "assistant"},
		{Path: fmt.Sprintf("entries.%04d.current_version", entry), Value: 0},
		{Path: fmt.Sprintf("entries.%04d.last_version", entry), Value: 0},
		{Path: fmt.Sprintf("entries.%04d.tokens_prompts", entry), Value: firestore.Increment(tokensPrompt)},
		{Path: fmt.Sprintf("entries.%04d.tokens_responses", entry), Value: firestore.Increment(tokensResponse)},
		{Path: fmt.Sprintf("entries.%04d.enabled", entry), Value: true},
		{Path: fmt.Sprintf("entries.%04d.versions.00.content", entry), Value: message},
		{Path: fmt.Sprintf("entries.%04d.versions.00.timestamp", entry), Value: time.Now()},
		{Path: fmt.Sprintf("entries.%04d.versions.00.tokens_prompt", entry), Value: tokensPrompt},
		{Path: fmt.Sprintf("entries.%04d.versions.00.tokens_response", entry), Value: tokensResponse},
		{Path: fmt.Sprintf("entries.%04d.versions.00.personality", entry), Value: m.Personality},
	}

	if _, err = c.Doc(model).Update(ctx, updates); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to update model document", "error", err)
		return err
	}

	// Update session

	path = fmt.Sprintf("sage/%s/projects/%s/sessions", username, project)

	c = client.Collection(path)
	if c == nil {
		return common.ErrNotFound{}
	}

	updates = []firestore.Update{
		{Path: fmt.Sprintf("models.%s.tokens_prompts", model), Value: firestore.Increment(tokensPrompt)},
		{Path: fmt.Sprintf("models.%s.tokens_responses", model), Value: firestore.Increment(tokensResponse)},
		{Path: fmt.Sprintf("models.%s.status", model), Value: "in progress"},
	}

	if _, err = c.Doc(session).Update(ctx, updates); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to update session document", "error", err)
		return err
	}

	// Update project

	path = fmt.Sprintf("sage/%s/projects", username)

	c = client.Collection(path)
	if c == nil {
		logCtx.Error("collection not found", "path", path)
		return common.ErrNotFound{}
	}

	modelEncoded := strings.ReplaceAll(m.Model, ".", "--")

	updates = []firestore.Update{
		{Path: fmt.Sprintf("tokens.%s_prompts", modelEncoded), Value: firestore.Increment(tokensPrompt)},
		{Path: fmt.Sprintf("tokens.%s_responses", modelEncoded), Value: firestore.Increment(tokensResponse)},
	}

	if _, err = c.Doc(project).Update(ctx, updates); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to update project document", "error", err)
		return err
	}

	// Update user

	c = client.Collection("sage")
	if c == nil {
		logCtx.Error("collection not found", "path", "sage")
		return common.ErrNotFound{}
	}
	if _, err = c.Doc(username).Update(ctx, updates); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to update user document", "error", err)
		return err
	}

	return nil
}

// CreateResponse creates a new Response type
func CreateResponse(ctx context.Context, logCtx *slog.Logger, username, project, session, model string) (Response, error) {
	logCtx = logCtx.With("fid", "sage.models.chatpad.CreateGPTBlock")

	client, err := firestore.NewClient(ctx, config.VARS.FirebaseProject)
	if err != nil {
		logCtx.Error("unable to create firestore client", "error", err)
		return Response{}, err
	}
	defer client.Close()

	m, err := GetGPT(ctx, logCtx, client, username, project, session, model)
	if err != nil {
		logCtx.Error("unable to get model", "error", err)
		return Response{}, err
	}

	if m.Entries[fmt.Sprintf("%04d", m.LastEntry)].Role != "assistant" {
		return Response{}, common.ErrNoResults
	}

	e, ok := m.Entries[fmt.Sprintf("%04d", m.LastEntry)]
	if !ok {
		return Response{}, common.ErrNoResults
	}

	v, ok := e.Versions[fmt.Sprintf("%02d", e.LastVersion)]
	if !ok {
		return Response{}, common.ErrNoResults
	}

	res := Response{
		LastEntry:             m.LastEntry,
		TokensVersionPrompt:   v.TokensPrompt,
		TokensVersionResponse: v.TokensResponse,
		TokensEntryPrompts:    e.TokensPrompts,
		TokensEntryResponses:  e.TokensResponses,
		TokensModelPrompts:    m.TokensPrompts,
		TokensModelResponses:  m.TokensResponses,
	}

	return res, nil
}

func addGPTPromptEntry(ctx context.Context, logCtx *slog.Logger, client *firestore.Client, username, project, session, model, message string) (int, error) {
	logCtx = logCtx.With("fid", "sage.models.chatpad.addGPTPromptEntry")

	m, err := GetGPT(ctx, logCtx, client, username, project, session, model)
	if err != nil {
		logCtx.Error("unable to get model", "error", err)
		return 0, err
	}

	path := fmt.Sprintf("sage/%s/projects/%s/sessions/%s/models", username, project, session)

	c := client.Collection(path)
	if c == nil {
		return 0, common.ErrNotFound{}
	}

	entry := m.LastEntry + 1

	updates := []firestore.Update{
		{Path: "last_entry", Value: entry},
		{Path: fmt.Sprintf("entries.%04d.role", entry), Value: "user"},
		{Path: fmt.Sprintf("entries.%04d.current_version", entry), Value: 0},
		{Path: fmt.Sprintf("entries.%04d.last_version", entry), Value: 0},
		{Path: fmt.Sprintf("entries.%04d.enabled", entry), Value: true},
		{Path: fmt.Sprintf("entries.%04d.versions.00.content", entry), Value: message},
		{Path: fmt.Sprintf("entries.%04d.versions.00.timestamp", entry), Value: time.Now()},
	}

	_, err = c.Doc(model).Update(ctx, updates)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to update model document", "error", err)
		return 0, err
	}

	return entry, nil
}

// createGPTReq assumes the current message has already been saved.
func createGPTReq(ctx context.Context, logCtx *slog.Logger, client *firestore.Client, username, project, session, model, modelName string) (openai.ChatRequest, int, error) {
	logCtx = logCtx.With("fid", "sage.models.chatpad.createGPTReq")

	m, err := GetGPT(ctx, logCtx, client, username, project, session, model)
	if err != nil {
		logCtx.Error("unable to get model", "error", err)
		return openai.ChatRequest{}, 0, err
	}

	req := openai.ChatRequest{
		Model:      modelName,
		Creativity: m.Creativity,
		MaxTokens:  m.MaxTokens,
	}

	// Chars start at 4 so it can be rounded up to the nearest token after dividing by 4
	chars := 4

	if m.Personality != "" {
		req.Messages = append(req.Messages, openai.ChatMessage{Role: "system", Content: m.Personality})
		chars += len(m.Personality) + 16
	}

	e, ok := m.Entries[fmt.Sprintf("%04d", m.LastEntry)]
	if !ok {
		logCtx.Error("consistency")
		return openai.ChatRequest{}, 0, common.ErrConsistency
	}

	v, ok := e.Versions[fmt.Sprintf("%02d", e.CurrentVersion)]
	if !ok {
		logCtx.Error("consistency")
		return openai.ChatRequest{}, 0, common.ErrConsistency
	}

	req.Messages = append(req.Messages, openai.ChatMessage{Role: "user", Content: v.Content})
	chars += len(v.Content) + 16

	return req, chars / 4, nil
}
