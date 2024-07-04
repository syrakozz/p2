package stt

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"disruptive/lib/deepgram"
	"disruptive/lib/openai"
)

// File transcribes an audio file.
func File(ctx context.Context, filename, language, engine string) error {
	logCtx := slog.With("filename", filename, "language", language, "engine", engine)

	f, err := os.Open(filename)
	if err != nil {
		logCtx.Error("unable to read file", "error", err)
		return err
	}
	defer f.Close()

	var text string

	if engine == "openai" {
		logCtx.Info("Transcribing")
		text, err = openai.PostTranscriptionsText(ctx, logCtx, f, "mp3", language)
		if err != nil {
			logCtx.Error("unable to transcribe", "error", err)
			return err
		}
	}

	if engine == "deepgram" {
		logCtx.Info("Transcribing")
		text, language, err = deepgram.PostTranscriptionsText(ctx, logCtx, f, language, "")
		if err != nil {
			logCtx.Error("unable to transcribe", "error", err)
			return err
		}
	}

	fmt.Println(language)
	fmt.Println(text)
	return nil
}
