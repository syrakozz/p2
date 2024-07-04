package stt

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"disruptive/lib/microphone"
	"disruptive/lib/openai"
)

// Record transcribes an audio file created using the default microphone.
func Record(ctx context.Context, language string, silence string) error {
	logCtx := slog.With()

	file, err := os.CreateTemp("", "vox*.mp3")
	if err != nil {
		logCtx.Error("unable to create temp file", "error", err)
		return err
	}

	filename := file.Name()
	file.Close()

	logCtx = logCtx.With("filename", file.Name(), "language", language, "silence", silence)
	logCtx.Info("Recording")

	if err := microphone.Record(ctx, logCtx, silence, filename); err != nil {
		logCtx.Error("unable to record", "error", err)
		return err
	}

	logCtx.Info("Transcribing")

	f, err := os.Open(filename)
	if err != nil {
		logCtx.Error("unable to read file", "error", err)
		return err
	}
	defer f.Close()

	text, err := openai.PostTranscriptionsText(ctx, logCtx, f, "mp3", language)
	if err != nil {
		logCtx.Error("unable to transcribe", "error", err)
		return err
	}

	fmt.Println(text)
	return nil
}
