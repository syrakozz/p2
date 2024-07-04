package play

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"text/template"

	"disruptive/lib/common"
	"disruptive/lib/openai"
)

type systemPromptInput struct {
	Character  string
	Age        string
	Attributes string
}

func response(ctx context.Context, logCtx *slog.Logger, text string, req *Request, sessionMemory *playSessionMemory) (string, error) {
	logCtx = logCtx.With("fid", "vox.play.response")
	logCtx.Info("Responding", "character", req.Character)

	c, ok := characters[req.Character]
	if !ok {
		logCtx.Error("invalid character", "character", req.Character)
		return "", errors.New("invalid character")
	}

	t, err := template.New("prompt").Parse(systemPromptTemplate)
	if err != nil {
		logCtx.Error("unable to create prompt template", "error", err)
		return "", err
	}

	input := systemPromptInput{
		Character:  c.LongName,
		Age:        req.Age,
		Attributes: c.SystemAttributes,
	}

	var systemPrompt strings.Builder
	if err := t.Execute(&systemPrompt, input); err != nil {
		logCtx.Error("unable to execute system prompt template", "error", err)
		return "", err
	}

	chatReq := openai.ChatRequest{
		Model:      "gpt-3.5-turbo",
		Messages:   []openai.ChatMessage{{Role: "system", Content: systemPrompt.String()}},
		Creativity: req.Creativity,
		MaxTokens:  req.Tokens,
	}

	start := len(sessionMemory.Entries) - req.SessionMemory
	if start < 0 {
		start = 0
	}

	for _, e := range sessionMemory.Entries[start:] {
		chatReq.Messages = append(chatReq.Messages, openai.ChatMessage{Role: "user", Content: e.User})
		chatReq.Messages = append(chatReq.Messages, openai.ChatMessage{Role: "assistant", Content: e.Assistant})
	}

	chatReq.Messages = append(chatReq.Messages, openai.ChatMessage{
		Role:    "user",
		Content: fmt.Sprintf("%s Please respond with exactly %d words for this response.  %s. Answer in %d words.", c.QuestionAttributes, req.Words, text, req.Words),
	})

	if req.Verbose {
		fmt.Println(common.MarshalIndent(chatReq))
	}

	if req.Verbose {
		n := 0
		for i := 0; i < len(chatReq.Messages); i++ {
			n += len(chatReq.Messages[i].Content)
		}
		pLen = n // mod var
	}

	chatRes, err := openai.PostChat(ctx, logCtx, chatReq)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logCtx.Warn("timeout", "error", err)
			return "", err
		}
		logCtx.Error("unable to get chat response", "error", err)
		return "", err
	}

	// Add new session memory
	sessionMemory.Entries = append(sessionMemory.Entries, playSessionMemoryEntry{
		User:      text,
		Assistant: chatRes.Text,
	})

	logCtx.Info("Chat response", "tokens", chatRes.Usage, "prompt", chatRes.UsagePrompt, "response", chatRes.UsageResponse)

	return chatRes.Text, nil
}
