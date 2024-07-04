// Package anthropic interfaces with the Anthropic APIs.
package anthropic

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"disruptive/lib/common"
)

// CompletionRequest is the API request structure.
type CompletionRequest struct {
	Prompt      string  `json:"prompt"`
	Model       string  `json:"model"`
	MaxTokens   int     `json:"max_tokens_to_sample"`
	Stream      bool    `json:"stream"`
	Temperature float64 `json:"temperature"`
}

// CompletionResponse is the API response structure.
type CompletionResponse struct {
	Completion string `json:"completion"`
	StopReason string `json:"stop_reason"`
}

// Completion sends chat props to Anthropic.
func Completion(ctx context.Context, logCtx *slog.Logger, req CompletionRequest) (CompletionResponse, error) {
	logCtx = logCtx.With("fid", "anthropic.Completion")

	res, err := Resty.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&CompletionResponse{}).
		Post(completeEndpoint)

	if err != nil {
		logCtx.Error("anthropic completion endpoint failed", "error", err)
		return CompletionResponse{}, err
	}

	if res.StatusCode() == http.StatusUnauthorized {
		logCtx.Error("anthropic unauthorized", "status", res.Status())
		return CompletionResponse{}, common.ErrUnauthorized
	}

	if res.StatusCode() != http.StatusOK {
		logCtx.Error("anthropic completion endpoint failed", "status", res.Status())
		return CompletionResponse{}, errors.New(res.Status())
	}

	return *res.Result().(*CompletionResponse), nil
}
