package memory

import (
	"context"
	"log/slog"

	"disruptive/lib/common"
	"disruptive/lib/pinecone"
)

// Stats returns the pinecode stats information.
func Stats(ctx context.Context) error {
	logCtx := slog.With("namespace", "all")

	res, err := pinecone.Stats(ctx, logCtx, nil)
	if err != nil {
		return err
	}

	common.P(res)
	logCtx.Info("stats")
	return nil
}
