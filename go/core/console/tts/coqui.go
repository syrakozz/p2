package tts

import (
	"context"
	"log/slog"
	"os"

	"disruptive/lib/coqui"
)

// Coqui calls the low-level TTS APIs.
func Coqui(ctx context.Context, inFilename, inText, outFilename, prompt, voice string) error {
	logCtx := slog.With()

	if inFilename != "" {
		b, err := os.ReadFile(inFilename)
		if err != nil {
			return err
		}

		inText = string(b)
	}

	req := coqui.Request{
		Prompt: prompt,
		Text:   inText,
		Voice:  voice,
	}

	coqui.TTSFile(ctx, logCtx, req, outFilename)
	return nil
}
