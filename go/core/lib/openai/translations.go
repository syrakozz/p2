package openai

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"disruptive/lib/common"
)

type postTranslationsResponse struct {
	Text string `json:"text"`
}

// PostTranslationsText translates audio to English
func PostTranslationsText(ctx context.Context, logCtx *slog.Logger, file io.Reader) (string, error) {
	fid := slog.String("fid", "openai.PostTranslationsText")

	if err := whisper1Limiter.Wait(ctx); err != nil {
		logCtx.Error("limiter wait failed", fid, "error", err)
		return "", err
	}

	// Override the Resty timeout if it exists.
	if timeout := ctx.Value(common.TimeoutKey); timeout != nil {
		Resty.SetTimeout(timeout.(time.Duration))
	}

	res, err := Resty.R().
		SetContext(ctx).
		SetFileReader("file", "audio.mp3", file).
		SetFormData(map[string]string{"model": "whisper-1"}).
		SetResult(&postTranslationsResponse{}).
		Post(translationsEndpoint)

	if err != nil {
		logCtx.Error("openai translations endpoint failed", fid, "error", err)
		return "", err
	}

	if res.StatusCode() == http.StatusUnauthorized {
		logCtx.Error("openai unauthorized", fid, "status", res.Status())
		return "", common.ErrUnauthorized
	}

	if res.StatusCode() != http.StatusOK {
		logCtx.Error("openai transcriptions endpoint failed", fid, "staus", res.Status())
		return "", errors.New(res.Status())
	}

	return res.Result().(*postTranslationsResponse).Text, nil
}
