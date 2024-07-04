package coqui

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"disruptive/lib/common"
)

// TTSStream calls the Coqui voice API to create a wav.
func TTSStream(ctx context.Context, logCtx *slog.Logger, req Request) (io.Reader, error) {
	logCtx = logCtx.With("fid", "coqui.TTSStream")

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
			return nil, err
		}

		request.Voice = v
		endpoint = e
	}

	if endpoint == sampleFileEndpoint && len(req.Text) > maxSampleLength {
		logCtx.Warn("text length too long", "len", len(req.Text))
		return nil, common.ErrLimit
	} else if (endpoint == sampleFileXTTSEndpoint || endpoint == samplePromptXTTSEndpoint) && len(req.Text) > maxSampleLength {
		logCtx.Warn("text length too long", "len", len(req.Text))
		return nil, common.ErrLimit
	}

	res, err := Resty.R().
		SetContext(ctx).
		SetBody(request).
		SetResult(&renderResponse{}).
		Post(endpoint)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logCtx.Warn("timeout", "error", err)
			return nil, err
		}
		logCtx.Error("coqui render endpoint failed", "error", err)
		return nil, err
	}

	if res.StatusCode() != http.StatusCreated {
		logCtx.Error("coqui render endpoint failed", "error", string(res.Body()), "status", res.Status())
		return nil, errors.New(res.Status())
	}

	renderRes := res.Result().(*renderResponse)

	if renderRes.AudioURL == nil || *renderRes.AudioURL == "" {
		logCtx.Error("missing audio url")
		return nil, errors.New("missing audio url")
	}

	res, err = Resty.R().
		SetContext(ctx).
		SetDoNotParseResponse(true).
		Get(*renderRes.AudioURL)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logCtx.Warn("timeout", "error", err)
			return nil, err
		}
		logCtx.Error("coqui get audio file endpoint failed", "error", err)
		return nil, err
	}

	if res.StatusCode() != http.StatusOK {
		logCtx.Error("coqui get audio file endpoint failed", "status", res.Status())
		return nil, errors.New(res.Status())
	}

	r, w := io.Pipe()

	go func() {
		b := res.RawBody()
		defer b.Close()
		defer w.Close()

		buf := make([]byte, 8196)

		for {
			n, err := b.Read(buf)
			if err != nil {
				if err == io.EOF {
					break
				}

				logCtx.Error("unable to read chunk", "error", err)
				return
			}

			if _, err := w.Write(buf[:n]); err != nil {
				logCtx.Error("unable to write to chunk pipe", "error", err)
				return
			}
		}
	}()

	return r, nil
}
