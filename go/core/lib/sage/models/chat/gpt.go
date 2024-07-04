package chat

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

// GPTBlock contains a single block in the database.
type GPTBlock struct {
	Start          int       `firestore:"start"`
	End            int       `firestore:"end"`
	Timestamp      time.Time `firestore:"timestamp"`
	Summary        string    `firestore:"summary"`
	TokensPrompt   int       `firestore:"tokens_prompt"`
	TokensResponse int       `firestore:"tokens_response"`
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
	LastBlock       int                 `firestore:"last_block"`
	TokensPrompts   int                 `firestore:"tokens_prompts"`
	TokensResponses int                 `firestore:"tokens_responses"`
	Entries         map[string]GPTEntry `firestore:"entries"`
	Blocks          map[string]GPTBlock `firestore:"blocks"`
}

// GetGPT retrieves a ChatGPT model.
func GetGPT(ctx context.Context, logCtx *slog.Logger, client *firestore.Client, username, project, session, model string) (GPTDocument, error) {
	logCtx = logCtx.With("fid", "sage.models.chat.GetGPT")

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
	username, project, session, model, message string, tokensPrompt, tokensResponse int, prompts []string) error {
	logCtx = logCtx.With("fid", "sage.models.chat.AddGPTAssistantEntry")

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
		{Path: fmt.Sprintf("entries.%04d.versions.00.prompts", entry), Value: prompts},
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

// AddGPTAssistantEntryVersion adds a new assistant entry.
func AddGPTAssistantEntryVersion(ctx context.Context, logCtx *slog.Logger,
	username, project, session, model, message string, tokensPrompt, tokensResponse int, prompts []string) error {
	logCtx = logCtx.With("fid", "sage.models.chat.AddGPTAssistantEntryVersion")

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

	e, ok := m.Entries[fmt.Sprintf("%04d", m.LastEntry)]
	if !ok {
		return common.ErrNotFound{}
	}

	version := e.LastVersion + 1

	updates := []firestore.Update{
		{Path: "tokens_prompts", Value: firestore.Increment(tokensPrompt)},
		{Path: "tokens_responses", Value: firestore.Increment(tokensResponse)},
		{Path: fmt.Sprintf("entries.%04d.current_version", m.LastEntry), Value: version},
		{Path: fmt.Sprintf("entries.%04d.last_version", m.LastEntry), Value: version},
		{Path: fmt.Sprintf("entries.%04d.tokens_prompts", m.LastEntry), Value: firestore.Increment(tokensPrompt)},
		{Path: fmt.Sprintf("entries.%04d.tokens_responses", m.LastEntry), Value: firestore.Increment(tokensResponse)},
		{Path: fmt.Sprintf("entries.%04d.versions.%02d.content", m.LastEntry, version), Value: message},
		{Path: fmt.Sprintf("entries.%04d.versions.%02d.timestamp", m.LastEntry, version), Value: time.Now()},
		{Path: fmt.Sprintf("entries.%04d.versions.%02d.tokens_prompt", m.LastEntry, version), Value: tokensPrompt},
		{Path: fmt.Sprintf("entries.%04d.versions.%02d.tokens_response", m.LastEntry, version), Value: tokensResponse},
		{Path: fmt.Sprintf("entries.%04d.versions.%02d.personality", m.LastEntry, version), Value: m.Personality},
		{Path: fmt.Sprintf("entries.%04d.versions.%02d.prompts", m.LastEntry, version), Value: prompts},
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

// CreateGPTBlock creates a new block and summary if the number of chars threshold is met.
func CreateGPTBlock(ctx context.Context, logCtx *slog.Logger, username, project, session, model string) (Response, error) {
	logCtx = logCtx.With("fid", "sage.models.chat.CreateGPTBlock")

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

	// Start variables

	startEntry := 0
	if m.LastBlock >= 0 {
		startEntry = m.Blocks[fmt.Sprintf("%04d", m.LastBlock)].End + 1
	}

	b := m.LastBlock + 1
	startBlockEntry := 0
	if b >= 1 {
		startBlockEntry = m.Blocks[fmt.Sprintf("%04d", b-1)].End + 1
	}

	startSummaryEntry := 0
	if b >= 2 {
		startSummaryEntry = m.Blocks[fmt.Sprintf("%04d", b-2)].End + 1
	}

	e, ok := m.Entries[fmt.Sprintf("%04d", m.LastEntry)]
	if !ok {
		return Response{}, common.ErrNoResults
	}

	v, ok := e.Versions[fmt.Sprintf("%02d", e.LastVersion)]
	if !ok {
		return Response{}, common.ErrNoResults
	}

	blockRes := Response{
		LastEntry:             m.LastEntry,
		TokensVersionPrompt:   v.TokensPrompt,
		TokensVersionResponse: v.TokensResponse,
		TokensEntryPrompts:    e.TokensPrompts,
		TokensEntryResponses:  e.TokensResponses,
		TokensModelPrompts:    m.TokensPrompts,
		TokensModelResponses:  m.TokensResponses,
		UnlockFrom:            startSummaryEntry,
	}

	// Calculate chars
	chars := 0

	for i := startEntry; i <= m.LastEntry; i++ {
		e := m.Entries[fmt.Sprintf("%04d", i)]
		chars += len(e.Versions[fmt.Sprintf("%02d", e.CurrentVersion)].Content)
	}

	// 1000 tokens * 4 chars/token
	if chars < 4000 {
		return blockRes, nil
	}

	path := fmt.Sprintf("sage/%s/projects/%s/sessions/%s/models", username, project, session)

	c := client.Collection(path)
	if c == nil {
		logCtx.Error("collection not found", "path", path)
		return Response{}, common.ErrNotFound{}
	}

	sb := strings.Builder{}
	sb.WriteString("Summarize in 200 words.\n")

	// Include previous summary
	if b >= 2 {
		sb.WriteString(m.Blocks[fmt.Sprintf("%04d", b-2)].Summary)
	}

	// Include entries
	for i := startSummaryEntry; i <= m.LastEntry; i++ {
		e := m.Entries[fmt.Sprintf("%04d", i)]
		if e.Role != "assistant" {
			continue
		}
		sb.WriteString(e.Versions[fmt.Sprintf("%02d", e.CurrentVersion)].Content)
	}

	req := openai.ChatRequest{
		Model:    "gpt-3.5-turbo",
		Messages: []openai.ChatMessage{{Role: "user", Content: sb.String()}},
	}

	chat, err := openai.PostChat(ctx, logCtx, req)
	if err != nil {
		logCtx.Error("unable to post openai chat summary.")
		return Response{}, err
	}

	// Create block

	updates := []firestore.Update{
		{Path: "last_block", Value: b},
		{Path: "tokens_prompts", Value: firestore.Increment(chat.UsagePrompt)},
		{Path: "tokens_responses", Value: firestore.Increment(chat.UsageResponse)},
		{Path: fmt.Sprintf("blocks.%04d.start", b), Value: startBlockEntry},
		{Path: fmt.Sprintf("blocks.%04d.end", b), Value: m.LastEntry},
		{Path: fmt.Sprintf("blocks.%04d.summary", b), Value: chat.Text},
		{Path: fmt.Sprintf("blocks.%04d.tokens_prompt", b), Value: chat.UsagePrompt},
		{Path: fmt.Sprintf("blocks.%04d.tokens_response", b), Value: chat.UsageResponse},
		{Path: fmt.Sprintf("blocks.%04d.timestamp", b), Value: time.Now()},
	}

	if _, err = c.Doc(model).Update(ctx, updates); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to update model document", "error", err)
		return Response{}, err
	}

	// Update session

	path = fmt.Sprintf("sage/%s/projects/%s/sessions", username, project)

	c = client.Collection(path)
	if c == nil {
		return Response{}, common.ErrNotFound{}
	}

	updates = []firestore.Update{
		{Path: fmt.Sprintf("models.%s.tokens_prompts", model), Value: firestore.Increment(chat.UsagePrompt)},
		{Path: fmt.Sprintf("models.%s.tokens_responses", model), Value: firestore.Increment(chat.UsageResponse)},
	}

	if _, err = c.Doc(session).Update(ctx, updates); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to update session document", "error", err)
		return Response{}, err
	}

	// Update project

	path = fmt.Sprintf("sage/%s/projects", username)

	c = client.Collection(path)
	if c == nil {
		logCtx.Error("collection not found", "path", path)
		return Response{}, common.ErrNotFound{}
	}

	modelEncoded := strings.ReplaceAll(m.Model, ".", "--")

	updates = []firestore.Update{
		{Path: fmt.Sprintf("tokens.%s_prompts", modelEncoded), Value: firestore.Increment(chat.UsagePrompt)},
		{Path: fmt.Sprintf("tokens.%s_responses", modelEncoded), Value: firestore.Increment(chat.UsageResponse)},
	}

	if _, err = c.Doc(project).Update(ctx, updates); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to update project document", "error", err)
		return Response{}, err
	}

	// Update user

	c = client.Collection("sage")
	if c == nil {
		logCtx.Error("collection not found", "path", "sage")
		return Response{}, common.ErrNotFound{}
	}
	if _, err = c.Doc(username).Update(ctx, updates); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to update user document", "error", err)
		return Response{}, err
	}

	// Update ChatResponse with summary info
	blockRes.TokensModelPrompts += chat.UsagePrompt
	blockRes.TokensModelResponses += chat.UsageResponse
	blockRes.TokensSummary = chat.UsagePrompt + chat.UsageResponse
	blockRes.UnlockFrom = m.Blocks[fmt.Sprintf("%04d", b-1)].End + 1

	return blockRes, nil
}

func addGPTPromptEntry(ctx context.Context, logCtx *slog.Logger, client *firestore.Client, username, project, session, model, message string) (int, error) {
	logCtx = logCtx.With("fid", "sage.models.chat.addGPTPromptEntry")

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
func createGPTReq(ctx context.Context, logCtx *slog.Logger, client *firestore.Client, username, project, session, model, modelName string, includeLastEntry bool) (openai.ChatRequest, int, []string, error) {
	logCtx = logCtx.With("fid", "sage.models.chat.createGPTReq")

	m, err := GetGPT(ctx, logCtx, client, username, project, session, model)
	if err != nil {
		logCtx.Error("unable to get model", "error", err)
		return openai.ChatRequest{}, 0, nil, err
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

	// Block 0 and 1 start with entry 0.
	// Block 2 starts the block shifting and incorporates a summary.
	startEntry := 0
	prompts := []string{}

	if m.LastBlock >= 1 {
		b, ok := m.Blocks[fmt.Sprintf("%04d", m.LastBlock-1)]
		if !ok {
			return openai.ChatRequest{}, 0, nil, common.ErrConsistency
		}

		startEntry = b.End + 1
		prompts = append(prompts, fmt.Sprintf("blocks.%04d.summary", m.LastBlock-1))
		req.Messages = append(req.Messages, openai.ChatMessage{Role: "assistant", Content: b.Summary})
	}

	lastEntry := m.LastEntry
	if !includeLastEntry {
		lastEntry--
	}

	for i := startEntry; i <= lastEntry; i++ {
		e, ok := m.Entries[fmt.Sprintf("%04d", i)]
		if !ok {
			logCtx.Error("consistency", "entry", i)
			return openai.ChatRequest{}, 0, nil, common.ErrConsistency
		}

		v, ok := e.Versions[fmt.Sprintf("%02d", e.CurrentVersion)]
		if !ok {
			logCtx.Error("consistency", "entry", i, "version", e.CurrentVersion)
			return openai.ChatRequest{}, 0, nil, common.ErrConsistency
		}

		prompts = append(prompts, fmt.Sprintf("entries.%04d.versions.%02d", i, e.CurrentVersion))
		req.Messages = append(req.Messages, openai.ChatMessage{Role: e.Role, Content: v.Content})
		chars += len(v.Content) + 16
	}

	return req, chars / 4, prompts, nil
}
