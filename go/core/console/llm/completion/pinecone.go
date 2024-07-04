package completion

import (
	"context"
	"errors"
	"log/slog"

	"disruptive/lib/common"
	"disruptive/lib/openai"
	"disruptive/lib/pinecone"
)

func processMemory(ctx context.Context, logCtx *slog.Logger, memory, query string, topK int, verbose bool) ([]pinecone.QueryResponse, error) {
	// Query Embedding
	embeddingsReq := openai.EmbeddingsRequest{
		Model: "text-embedding-ada-002",
		Input: []string{query},
	}

	embeddingsRes, err := openai.PostEmbeddings(ctx, logCtx, embeddingsReq)
	if err != nil {
		logCtx.Error("unable to create embedding", "error", err)
		return nil, err
	}

	if len(embeddingsRes.Data) != 1 {
		logCtx.Error("invalid embedding data", "error", err)
		return nil, errors.New("invalid embeddings data")
	}

	// Query
	queryReq := pinecone.QueryRequest{
		Namespace:       memory,
		TopK:            topK,
		Vector:          embeddingsRes.Data[0].Embedding,
		IncludeMetadata: true,
	}

	queryRes, err := pinecone.Query(ctx, logCtx, queryReq)
	if err != nil {
		logCtx.Error("unable to query", "error", err)
		return nil, err
	}

	if len(queryRes) < 1 {
		logCtx.Error("no context found")
		return nil, errors.New("no context found")
	}

	if verbose {
		common.P(queryRes)
	}

	return queryRes, nil
}
