package tts

import (
	"context"
	"io"
	"log/slog"
	"os"

	"disruptive/lib/elevenlabs"
)

// Elevenlabs calls the low-level TTS APIs.
func Elevenlabs(ctx context.Context, inFilename, inText, outFilename, voice, language string) error {
	logCtx := slog.With()

	if inFilename != "" {
		b, err := os.ReadFile(inFilename)
		if err != nil {
			logCtx.Error("unable to read input file", "error", err)
			return err
		}

		inText = string(b)
	}

	req := elevenlabs.Request{
		Filename: outFilename,
		Format:   "mp3_44100_128",
		Voice:    voice,
		Text:     inText,
		Language: language,
	}

	if outFilename != "-" {
		return elevenlabs.TTSFile(ctx, logCtx, req)
	}

	r, err := elevenlabs.TTSStream(ctx, logCtx, req)
	if err != nil {
		logCtx.Error("unable to stream", "error", err)
		return err
	}

	if _, err := io.Copy(os.Stdout, r); err != nil {
		logCtx.Error("unable to copy stream bytes", "error", err)
		return err
	}

	return nil
}
