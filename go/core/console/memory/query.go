package memory

import (
	"context"
	"errors"
	"log/slog"

	"disruptive/lib/common"
	"disruptive/lib/openai"
	"disruptive/lib/pinecone"
)

// Query creates an embedding from a query and finds the topK closest documents in the vector database.
func Query(ctx context.Context, namespace, query string, topK int) error {
	logCtx := slog.With("namespace", namespace, "topK", topK, "query", query)

	embeddingReq := openai.EmbeddingsRequest{
		Model: "text-embedding-ada-002",
		Input: []string{query},
	}

	embeddingRes, err := openai.PostEmbeddings(ctx, logCtx, embeddingReq)
	if err != nil {
		logCtx.Error("unable to post embeddings", "error", err)
		return err
	}

	if len(embeddingRes.Data) != 1 {
		logCtx.Error("inconsistent query embedding data length", "len", len(embeddingRes.Data))
		return errors.New("inconsistent query embedding data length")
	}

	queryReq := pinecone.QueryRequest{
		Namespace:       namespace,
		TopK:            topK,
		Vector:          embeddingRes.Data[0].Embedding,
		IncludeMetadata: true,
	}

	queryRes, err := pinecone.Query(ctx, logCtx, queryReq)
	if err != nil {
		return err
	}

	common.P(queryRes)
	logCtx.Info("query")
	return nil
}
