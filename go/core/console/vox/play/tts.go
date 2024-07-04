package play

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"disruptive/lib/coqui"
	"disruptive/lib/elevenlabs"
)

func tts(ctx context.Context, logCtx *slog.Logger, req *Request, ttsFile *os.File, text string) error {
	logCtx = logCtx.With("fid", "vox.play.tts")

	c, ok := characters[req.Character]
	if !ok {
		logCtx.Error("invalid character")
		return errors.New("invalid character")
	}

	switch c.TTS {
	case "coqui":
		req := coqui.Request{
			Text:  text,
			Voice: c.Voice,
		}

		if err := coqui.TTSFile(ctx, logCtx, req, ttsFile.Name()); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				logCtx.Warn("timeout", "error", err)
				return err
			}
			logCtx.Error("invalid voice", "error", err, "voice", c.Voice)
			return fmt.Errorf("invalid voice: %w", err)
		}
	case "11labs":
		req := elevenlabs.Request{
			Filename:          ttsFile.Name(),
			Format:            "mp3_44100_128",
			Language:          "",
			Voice:             c.Voice,
			Text:              text,
			SimilarityBoost:   req.VoiceSimilarityBoost,
			Stability:         req.VoiceStability,
			StyleExaggeration: req.VoiceStyleExaggeration,
		}

		if err := elevenlabs.TTSFile(ctx, logCtx, req); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				logCtx.Warn("timeout", "error", err)
				return err
			}
			logCtx.Error("invalid voice", "error", err, "voice", c.Voice)
			return fmt.Errorf("invalid voice: %w", err)
		}
	default:
		logCtx.Error("invalid tts", "tts", c.TTS)
		return errors.New("invalid tts")
	}

	return nil
}
