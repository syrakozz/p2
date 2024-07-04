package memory

import (
	"context"
	"log/slog"

	"disruptive/lib/mbox"
)

// ProcessMboxConsole process mbox files to console.
func ProcessMboxConsole(ctx context.Context, filename, namespace string, continueFlag bool) error {
	logCtx := slog.With("filename", filename, "namespace", namespace, "continue", continueFlag)

	_, err := mbox.Process(ctx, logCtx, filename, namespace, mbox.ConsoleFunc, continueFlag)
	return err
}

// ProcessMboxPinecone process mbox files.
func ProcessMboxPinecone(ctx context.Context, filename, namespace string, continueFlag bool) error {
	logCtx := slog.With("filename", filename, "namespace", namespace, "continueFlag", continueFlag)

	_, err := mbox.Process(ctx, logCtx, filename, namespace, mbox.PineconeFunc, continueFlag)
	return err
}
