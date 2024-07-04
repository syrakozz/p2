package play

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"disruptive/config"
	"disruptive/lib/common"
	"disruptive/lib/configs"
	"disruptive/lib/openai"
	"disruptive/lib/stabilityai"
	"disruptive/pkg/vox/characters"
	"disruptive/pkg/vox/profiles"
)

const (
	ttiPrompt = `Summarize and create a concise and accurate Stable Diffusion text-to-image prompt for the following text in triple quotes.	'''%s'''`
)

// TTIRequest contains the highlevel request to the Stability AI API.
type TTIRequest struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// TTI creates a text-to-image.
func TTI(ctx context.Context, logCtx *slog.Logger, profileID, characterVersion, engine string, width, height, n int) ([]byte, error) {
	fid := slog.String("fid", "tti.TTI")
	t := time.Now()

	p, err := profiles.GetByID(ctx, logCtx, profileID)
	if err != nil {
		logCtx.Error("unable to get profile", fid, "error", err)
		return nil, err
	}

	characterName := strings.Split(characterVersion, "_")[0]

	s, err := characters.GetSession(ctx, logCtx, profileID, characterName)
	if err != nil {
		logCtx.Error("unable to get session", fid, "error", err)
		return nil, err
	}

	text, err := getTTITextFromSession(ctx, logCtx, &s, n)
	if text == "" {
		return nil, common.ErrNoResults
	}

	style := p.Characters[characterName].ImageStyle
	if style == "" {
		style = "comic-book"
	}

	if width%64 != 0 {
		logCtx.Error("invalid image size", "widgth", width)
		return nil, common.ErrBadRequest{Src: "tti", Msg: "invalid image width"}
	}

	if height%64 != 0 {
		logCtx.Error("invalid image size", "height", height)
		return nil, common.ErrBadRequest{Src: "tti", Msg: "invalid image height"}
	}

	stabilityReq := stabilityai.Request{
		StylePreset: style,
		Width:       width,
		Height:      height,
	}

	stabilityReq.TextPrompts = append(stabilityReq.TextPrompts, stabilityai.TextPrompt{
		Text:   text,
		Weight: 0.5,
	})

	config, err := configs.GetStableDiffusion(ctx, logCtx)
	if err != nil {
		logCtx.Error("unable to get stablediffusion config", fid, "error", err)
		return nil, err
	}

	stabilityReq.TextPrompts = append(stabilityReq.TextPrompts, stabilityai.TextPrompt{
		Text:   strings.Join(config.NegativePrompts, ", "),
		Weight: -1,
	})

	image, err := stabilityai.PostTTI(ctx, logCtx, engine, stabilityReq)
	if err != nil {
		logCtx.Error("unable to create text-to-image", fid, "error", err)
		return nil, err
	}

	logCtx.Info("duration", "duration", time.Since(t).Milliseconds(), "span", "tti")

	return image, nil
}

func getTTITextFromSession(ctx context.Context, logCtx *slog.Logger, session *characters.SessionDocument, n int) (string, error) {
	fid := slog.String("fid", "tti.getTTITextFromSession")

	var entries strings.Builder

	start := len(session.Entries) + session.StartEntry - n
	if start < 1 {
		start = session.StartEntry
	}

	for i := start; i < start+n; i++ {
		entries.WriteString(session.Entries[fmt.Sprintf("%06d", i)].Assistant)
	}

	chatReq := openai.ChatRequest{
		Model:     "gpt-3.5-turbo",
		Messages:  []openai.ChatMessage{{Role: "user", Content: fmt.Sprintf(ttiPrompt, entries.String())}},
		MaxTokens: 250,
	}

	chatRes, err := openai.PostChat(ctx, logCtx, chatReq)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logCtx.Error("timeout", fid, "error", err)
			return "", err
		}
		logCtx.Error("unable to get chat response", fid, "error", err)
		return "", err
	}

	promptCost := float64(chatRes.UsagePrompt) * config.VARS.GPT35TurboPromptCost / 1000.0
	responseCost := float64(chatRes.UsageResponse) * config.VARS.GPT35TurboResponseCost / 1000.0

	logCtx.Info(
		"cost",
		"feature", "tti",
		"model", "gpt-3.5-turbo",
		"tokens_prompt", chatRes.UsagePrompt,
		"tokens_response", chatRes.UsageResponse,
		"tokens_total", chatRes.UsagePrompt+chatRes.UsageResponse,
		"cost_prompt", fmt.Sprintf("%.7f", promptCost),
		"cost_response", fmt.Sprintf("%.7f", responseCost),
		"cost_total", fmt.Sprintf("%.7f", promptCost+responseCost),
	)

	return chatRes.Text, nil
}
