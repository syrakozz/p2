package elevenlabs

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"disruptive/lib/common"
)

// TTSFile calls the ElevenLabs voice API to create an mp3.
func TTSFile(ctx context.Context, logCtx *slog.Logger, req Request) error {
	fid := slog.String("fid", "elevenlabs.TTSFile")

	if req.OptimizeStreamingLatency == "" {
		req.OptimizeStreamingLatency = "3"
	}

	v, ok := Voices[req.Voice]
	if !ok {
		logCtx.Error("voice not found", fid, "voice", req.Voice)
		return common.ErrNotFound{Msg: "voice not found"}
	}

	res, err := Resty.R().
		SetContext(ctx).
		SetHeader("Accept", AudioFormatContentTypes[req.Format]).
		SetBody(renderFromRequest(req)).
		SetPathParam("voice", v.ID).
		SetOutput(req.Filename).
		SetQueryParam("optimize_streaming_latency", req.OptimizeStreamingLatency).
		SetQueryParam("output_format", req.Format).
		Post(voiceElevenLabsEndpoint)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logCtx.Error("timeout", fid, "error", err)
			return err
		}
		logCtx.Error("elevenlabs voice endpoint failed", fid, "error", err)
		return err
	}

	if res.StatusCode() != http.StatusOK {
		logCtx.Error("elevenlabs voice endpoint failed", fid, "status", res.Status())
		return errors.New(res.Status())
	}

	return nil
}
