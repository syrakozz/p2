package play

import (
	"context"
	"log/slog"
	"os"

	"disruptive/lib/openai"
)

func stt(ctx context.Context, logCtx *slog.Logger, filename, language string) (string, error) {
	logCtx = logCtx.With("fid", "vox.play.stt")

	f, err := os.Open(filename)
	if err != nil {
		logCtx.Error("unable to read file", "error", err)
		return "", err
	}
	defer f.Close()

	text, err := openai.PostTranscriptionsText(ctx, logCtx, f, "mp3", language)
	if err != nil {
		logCtx.Error("unable to transcribe", "error", err)
		return "", err
	}

	return text, err
}
