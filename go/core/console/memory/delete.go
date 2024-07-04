package memory

import (
	"context"
	"log/slog"

	"disruptive/lib/pinecone"
)

// Delete memory by IDs or all within a single namespace.
func Delete(ctx context.Context, namespace string, ids []string, deleteAll bool) error {
	logCtx := slog.With("num_ids", len(ids), "deleteAll", deleteAll)

	request := pinecone.DeleteRequest{
		Namespace: namespace,
		IDs:       ids,
		DeleteAll: deleteAll,
	}

	if err := pinecone.Delete(ctx, logCtx, request); err != nil {
		logCtx.Error("unable to delete", "error", err)
		return err
	}

	logCtx.Info("deleted")
	return nil
}
