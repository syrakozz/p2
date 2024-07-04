package stabilityai

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
)

type (
	// Request contains in the input request parameters.
	Request struct {
		TextPrompts        []TextPrompt `json:"text_prompts"`
		Height             int          `json:"height,omitempty"`
		Width              int          `json:"width,omitempty"`
		CfgScale           int          `json:"cfg_scare,omitempty"`
		ClipGuidancePreset string       `json:"clip_guidance_preset,omitempty"`
		Sampler            string       `json:"sampler,omitempty"`
		Samples            int          `json:"samples,omitempty"`
		Seed               int          `json:"seed,omitempty"`
		Steps              int          `json:"steps,omitempty"`
		StylePreset        string       `json:"style_preset,omitempty"`
	}

	// TextPrompt is one part of an image.
	TextPrompt struct {
		Text   string  `json:"text"`
		Weight float64 `json:"weight"`
	}
)

// PostTTI calls the StabilityAI text-to-image API.
func PostTTI(ctx context.Context, logCtx *slog.Logger, engine string, req Request) ([]byte, error) {
	fid := slog.String("fid", "stabilityai.PostTTI")

	engineID, ok := Engines[engine]
	if !ok {
		engineID = "stable-diffusion-v1-6"
	}

	res, err := Resty.R().
		SetContext(ctx).
		SetHeader("Accept", "image/png").
		SetPathParam("engine_id", engineID).
		SetBody(req).
		Post(textToImageEndpoint)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logCtx.Error("timeout", fid, "error", err)
			return nil, err
		}
		logCtx.Error("stabilityai text-to-image endpoint failed", fid, "error", err)
		return nil, err
	}

	if res.StatusCode() != http.StatusOK {
		logCtx.Error("stabilityai text-to-image endpoint failed", fid, "status", res.Status())
		return nil, errors.New(res.Status())
	}

	return res.Body(), nil
}
