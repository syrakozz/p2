package chatpad

import (
	"context"
	"io"
	"log/slog"

	"cloud.google.com/go/firestore"

	"disruptive/config"
	"disruptive/lib/common"
	"disruptive/lib/openai"
	"disruptive/lib/sage/sessions"
)

// Response contains finalize information.
// These are value that are not convenient to return to the UI after a streaming chat response.
type Response struct {
	Text                  string `json:"text,omitempty"`
	FinishReason          string `json:"finish_reason,omitempty"`
	LastEntry             int    `json:"last_entry"`
	TokensVersionPrompt   int    `json:"tokens_version_prompt"`
	TokensVersionResponse int    `json:"tokens_version_response"`
	TokensEntryPrompts    int    `json:"tokens_entry_prompts"`
	TokensEntryResponses  int    `json:"tokens_entry_responses"`
	TokensModelPrompts    int    `json:"tokens_model_prompts"`
	TokensModelResponses  int    `json:"tokens_model_responses"`
}

// Post returns an io.reader which is filled one token at at time.
func Post(ctx context.Context, logCtx *slog.Logger, username, project, session, model, prompt string) (Response, error) {
	logCtx = logCtx.With("fid", "sage.models.chatpad.Post")

	client, err := firestore.NewClient(ctx, config.VARS.FirebaseProject)
	if err != nil {
		logCtx.Error("unable to create firestore client", "error", err)
		return Response{}, err
	}
	defer client.Close()

	s, err := sessions.Get(ctx, logCtx, username, project, session)
	if err != nil {
		logCtx.Error("unable to get session", "error", err)
		return Response{}, err
	}

	m, ok := s.Models[model]
	if !ok {
		logCtx.Error("unable to get model info")
		return Response{}, common.ErrNotFound{Msg: "model not found"}
	}

	if !(m.Type == "chatpad" && m.Model == "gpt-3.5-turbo") {
		return Response{}, common.ErrBadRequest{Msg: "invalid chat type and model"}
	}

	if _, err := addGPTPromptEntry(ctx, logCtx, client, username, project, session, model, prompt); err != nil {
		logCtx.Error("unable to add prompt entry")
		return Response{}, err
	}

	req, _, err := createGPTReq(ctx, logCtx, client, username, project, session, model, m.Model)
	if err != nil {
		logCtx.Error("unable to create chat request")
		return Response{}, err
	}

	chatRes, err := openai.PostChat(ctx, logCtx, req)
	if err != nil {
		logCtx.Error("unable to post openai chat.")
		return Response{}, err
	}

	if err := AddGPTAssistantEntry(ctx, logCtx, username, project, session, model, chatRes.Text, chatRes.UsagePrompt, chatRes.UsageResponse); err != nil {
		logCtx.Error("unable to add response entry")
		return Response{}, err
	}

	res, err := CreateResponse(ctx, logCtx, username, project, session, model)
	if err != nil {
		logCtx.Error("unable to create response")
		return Response{}, err
	}

	res.Text = chatRes.Text
	res.FinishReason = chatRes.FinishReason
	return res, nil
}

// PostStream returns an io.reader which is filled one token at at time.
func PostStream(ctx context.Context, logCtx *slog.Logger, username, project, session, model, prompt string) (io.Reader, int, int, error) {
	logCtx = logCtx.With("fid", "sage.models.chatpad.PostStream")

	client, err := firestore.NewClient(ctx, config.VARS.FirebaseProject)
	if err != nil {
		logCtx.Error("unable to create firestore client", "error", err)
		return nil, 0, 0, err
	}
	defer client.Close()

	s, err := sessions.Get(ctx, logCtx, username, project, session)
	if err != nil {
		logCtx.Error("unable to get session", "error", err)
		return nil, 0, 0, err
	}

	m, ok := s.Models[model]
	if !ok {
		logCtx.Error("unable to get model info")
		return nil, 0, 0, common.ErrNotFound{Msg: "model not found"}
	}

	if !(m.Type == "chatpad" && m.Model == "gpt-3.5-turbo") {
		return nil, 0, 0, common.ErrBadRequest{Msg: "invalid chat type and model"}
	}

	entry, err := addGPTPromptEntry(ctx, logCtx, client, username, project, session, model, prompt)
	if err != nil {
		return nil, 0, 0, err
	}

	req, tokens, err := createGPTReq(ctx, logCtx, client, username, project, session, model, m.Model)
	if err != nil {
		return nil, 0, 0, err
	}

	r, err := openai.PostChatStream(ctx, logCtx, req)

	return r, tokens, entry, err
}
