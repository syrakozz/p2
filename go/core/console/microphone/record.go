// Package microphone uses the microphone to record a wave file.
package microphone

import (
	"context"
	"log/slog"

	"disruptive/lib/microphone"
)

// Record records a wave file using the default microphone.
func Record(ctx context.Context, silence, filename string) error {
	logCtx := slog.With("filename", filename, "silence", silence)
	logCtx.Info("Recording")
	return microphone.Record(ctx, logCtx, silence, filename)
}
