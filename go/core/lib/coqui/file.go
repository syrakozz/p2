package coqui

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"disruptive/lib/common"
)

// TTSFile calls the Coqui voice API to create a wav.
func TTSFile(ctx context.Context, logCtx *slog.Logger, req Request, filename string) error {
	logCtx = logCtx.With("fid", "coqui.TTSFile")

	request := renderRequest{
		Speed: 1,
		Text:  req.Text,
	}

	var endpoint string

	if req.Prompt != "" {
		request.Prompt = req.Prompt
		endpoint = samplePromptXTTSEndpoint
	} else {
		v, e, err := getVoice(req.Voice)
		if err != nil {
			logCtx.Error("invalid voice", "error", err)
			return err
		}

		request.Voice = v
		endpoint = e
	}

	if endpoint == sampleFileEndpoint && len(req.Text) > maxSampleLength {
		logCtx.Warn("text length too long", "len", len(req.Text))
		return common.ErrLimit
	} else if (endpoint == sampleFileXTTSEndpoint || endpoint == samplePromptXTTSEndpoint) && len(req.Text) > maxSampleLength {
		logCtx.Warn("text length too long", "len", len(req.Text))
		return common.ErrLimit
	}

	timeGenerate := time.Now()

	res, err := Resty.R().
		SetContext(ctx).
		SetBody(request).
		SetResult(&renderResponse{}).
		Post(endpoint)

	sinceGenerate := time.Since(timeGenerate)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logCtx.Warn("timeout", "error", err)
			return err
		}
		logCtx.Error("coqui render endpoint failed", "error", err)
		return err
	}

	if res.StatusCode() != http.StatusCreated {
		logCtx.Error("coqui render endpoint failed", "error", string(res.Body()), "status", res.Status())
		return errors.New(res.Status())
	}

	renderRes := res.Result().(*renderResponse)

	if renderRes.AudioURL == nil || *renderRes.AudioURL == "" {
		logCtx.Error("missing audio url")
		return errors.New("missing audio url")
	}

	timeDownload := time.Now()

	res, err = Resty.R().
		SetContext(ctx).
		SetOutput(filename).
		Get(*renderRes.AudioURL)

	sinceDownload := time.Since(timeDownload)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logCtx.Warn("timeout", "error", err)
			return err
		}
		logCtx.Error("coqui get audio file endpoint failed", "error", err)
		return err
	}

	if res.StatusCode() != http.StatusOK {
		logCtx.Error("coqui get audio file endpoint failed", "status", res.Status())
		return errors.New(res.Status())
	}

	logCtx.Info("time", "generate", sinceGenerate, "download", sinceDownload, "total", sinceGenerate+sinceDownload)

	return nil
}
