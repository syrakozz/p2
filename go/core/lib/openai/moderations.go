package openai

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"disruptive/lib/common"
)

// ModerationResponse contains the moderation categories and scores.
type ModerationResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Results []struct {
		Categories     map[string]bool    `json:"categories"`
		CategoryScores map[string]float64 `json:"category_scores"`
	} `json:"results"`
}

// PostModeration returns moderation scores for some text input.
func PostModeration(ctx context.Context, logCtx *slog.Logger, text string) (ModerationResponse, error) {
	fid := slog.String("fid", "openai.Moderations")

	if text == "" {
		logCtx.Warn("input must be length of 2")
		return ModerationResponse{}, common.ErrBadRequest{Msg: "missing text"}
	}

	res, err := Resty.R().
		SetContext(ctx).
		SetBody(map[string]string{"input": text}).
		SetResult(&ModerationResponse{}).
		Post(moderationsEndpoint)

	if err != nil {
		logCtx.Error("openai moderations endpoint failed", fid, "error", err)
		return ModerationResponse{}, err
	}

	if res.StatusCode() == http.StatusUnauthorized {
		logCtx.Error("openai unauthorized", fid, "status", res.Status())
		return ModerationResponse{}, common.ErrUnauthorized
	}

	if res.StatusCode() != http.StatusOK {
		logCtx.Error("openai moderations endpoint failed", fid, "status", res.Status(), "message", errorMessage(res.Body()))
		return ModerationResponse{}, errors.New(res.Status())
	}

	return *res.Result().(*ModerationResponse), nil
}
