package translate

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"disruptive/lib/openai"
)

// File translates an audio file to English.
func File(ctx context.Context, filename string) error {
	logCtx := slog.With("filename", filename)

	logCtx.Info("Translating")

	f, err := os.Open(filename)
	if err != nil {
		logCtx.Error("unable to read file", "error", err)
		return err
	}

	text, err := openai.PostTranslationsText(ctx, logCtx, f)
	if err != nil {
		logCtx.Error("unable to transcribe", "error", err)
		return err
	}

	fmt.Println(text)
	return nil
}
