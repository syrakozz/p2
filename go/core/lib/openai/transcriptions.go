package openai

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	iso "github.com/emvi/iso-639-1"

	"disruptive/lib/common"
)

type postTranscriptionsResponse struct {
	Text string `json:"text"`
}

// PostTranscriptionsText sends transcriptions to OpenAI and returns the text.
// language is ISO-639-1 in format: https://en.wikipedia.org/wiki/List_of_ISO_639-1_codes
func PostTranscriptionsText(ctx context.Context, logCtx *slog.Logger, file io.Reader, extension, language string) (string, error) {
	fid := slog.String("fid", "openai.PostTranscriptionsText")

	if language != "" && !iso.ValidCode(language) {
		logCtx.Error("invalid language", "language", language)
		return "", errors.New("invalid language")
	}

	if err := whisper1Limiter.Wait(ctx); err != nil {
		logCtx.Error("limiter wait failed", fid, "error", err)
		return "", err
	}

	// Override the Resty timeout if it exists.
	if timeout := ctx.Value(common.TimeoutKey); timeout != nil {
		Resty.SetTimeout(timeout.(time.Duration))
	}

	req := map[string]string{"model": "whisper-1"}

	if language != "" {
		req["language"] = language
	}

	res, err := Resty.R().
		SetContext(ctx).
		SetFileReader("file", "audio."+extension, file).
		SetFormData(req).
		SetResult(&postTranscriptionsResponse{}).
		Post(transcriptionsEndpoint)

	if err != nil {
		logCtx.Error("openai transcriptions endpoint failed", fid, "error", err)
		return "", err
	}

	if res.StatusCode() == http.StatusUnauthorized {
		logCtx.Error("openai unauthorized", fid, "status", res.Status())
		return "", common.ErrUnauthorized
	}

	if res.StatusCode() != http.StatusOK {
		logCtx.Error("openai transcriptions endpoint failed", fid, "status", res.Status())
		return "", errors.New(res.Status())
	}

	return res.Result().(*postTranscriptionsResponse).Text, nil
}
