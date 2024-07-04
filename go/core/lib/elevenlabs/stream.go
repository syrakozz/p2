package elevenlabs

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"disruptive/lib/opus"
)

// TTSStream calls the ElevenLabs voice API to create an mp3.
func TTSStream(ctx context.Context, logCtx *slog.Logger, req Request) (io.Reader, error) {
	fid := slog.String("fid", "elevenlabs.TTSStream")

	if req.OptimizeStreamingLatency == "" {
		req.OptimizeStreamingLatency = "0"
	}

	var isOpus bool

	if req.Format == "opus_16000" {
		isOpus = true
		req.Format = "pcm_16000"
	}

	res, err := Resty.R().
		SetContext(ctx).
		SetDoNotParseResponse(true).
		SetHeader("Accept", AudioFormatContentTypes[req.Format]).
		SetBody(renderFromRequest(req)).
		SetPathParam("voice", req.Voice).
		SetQueryParam("optimize_streaming_latency", req.OptimizeStreamingLatency).
		SetQueryParam("output_format", req.Format).
		Post(voiceElevenLabsStreamEndpoint)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logCtx.Error("timeout", "error", fid, err)
			return nil, err
		}
		logCtx.Error("elevenlabs voice stream endpoint failed", fid, "error", err)
		return nil, err
	}

	if res.StatusCode() != http.StatusOK {
		logCtx.Error("elevenlabs voice stream endpoint failed", fid, "status", res.Status())
		return nil, errors.New(res.Status())
	}

	r, w := io.Pipe()

	if isOpus {
		go func() {
			b := res.RawBody()
			defer b.Close()
			defer w.Close()

			opusEncoder, err := opus.CreateOpusEncoder()
			if err != nil {
				logCtx.Error("unable to create opus encoder", fid, "error", err)
				return
			}
			defer opus.DestroyOpusEncoder(opusEncoder)

			if err := opus.InitOpusEncoder(opusEncoder); err != nil {
				logCtx.Error("unable to set constant bitrate", fid, "error", err)
				return
			}

			buf := make([]byte, 16384)
			acc := make([]byte, 0)

			for {
				select {
				case <-ctx.Done():
					logCtx.Info("canceled")
					return
				default:
				}

				n, err := b.Read(buf)
				if n > 0 {
					acc = append(acc, buf[:n]...)

					frames := len(acc) / opus.FrameSize
					if frames < 1 {
						continue
					}

					for i := frames; i > 0; i-- {
						goData := acc[:opus.FrameSize]
						acc = acc[opus.FrameSize:]

						if err := opus.WriteOpusFrame(opusEncoder, w, goData); err != nil {
							logCtx.Error("unable to encode opus frame", fid, "error", err)
							return
						}
					}
				}

				if err != nil {
					if err == io.EOF {
						break
					}

					logCtx.Warn("unable to read chunk", fid, "error", err)
					return
				}

			}

			if len(acc) > 0 {
				zeroBytes := make([]byte, opus.FrameSize-len(acc))
				acc = append(acc, zeroBytes...)
				if err := opus.WriteOpusFrame(opusEncoder, w, acc); err != nil {
					logCtx.Error("unable to encode opus frame", fid, "error", err)
					return
				}
			}
		}()
	} else {
		go func() {
			b := res.RawBody()
			defer b.Close()
			defer w.Close()

			buf := make([]byte, 16384)

			for {
				n, err := b.Read(buf)
				if n > 0 {
					if _, err := w.Write(buf[:n]); err != nil {
						logCtx.Error("unable to write to chunk pipe", fid, "error", err)
						return
					}
				}

				if err != nil {
					if err == io.EOF {
						break
					}

					logCtx.Warn("unable to read chunk", fid, "error", err)
					return
				}

			}
		}()
	}
	return r, nil
}
