package stress

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"golang.org/x/sync/errgroup"

	"disruptive/lib/openai"
)

// Embeddings stress tests the OpenAI embedding API
func Embeddings(ctx context.Context, text string, num int) error {
	logCtx := slog.With("num", num)

	g, ctx := errgroup.WithContext(ctx)

	results.Num = num
	results.Durations = make([]time.Duration, 0, num)

	t := time.Now()
	for i := 0; i < num; i++ {
		g.Go(func() error {
			return embeddings(ctx, logCtx, text)
		})
	}

	if err := g.Wait(); err != nil {
		if !errors.Is(err, context.Canceled) {
			logCtx.Error("unable to process all embeddings", "error", err)
		}
	}

	results.Total = time.Since(t)
	results.TotalS = results.Total.String()

	printResults(&results)
	return nil
}

func embeddings(ctx context.Context, logCtx *slog.Logger, text string) error {
	embeddingsReq := openai.EmbeddingsRequest{
		Model: "text-embedding-ada-002",
		Input: []string{text},
	}

	t := time.Now()

	embeddingsRes, err := openai.PostEmbeddings(ctx, logCtx, embeddingsReq)
	if err != nil {
		logCtx.Error("unable to create embedding", "error", err)
		return err
	}

	d := time.Since(t)

	if len(embeddingsRes.Data) != 1 {
		logCtx.Error("invalid embedding data", "error", err)
		return errors.New("invalid embeddings data")
	}

	results.Lock()
	results.Durations = append(results.Durations, d)
	results.Unlock()
	return nil
}
