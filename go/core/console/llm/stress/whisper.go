// Package stress ...
package stress

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"golang.org/x/sync/errgroup"

	"disruptive/lib/openai"
)

// Whisper stress tests the OpenAI embedding API
func Whisper(ctx context.Context, file string, num int) error {
	logCtx := slog.With("file", file, "num", num)

	g, ctx := errgroup.WithContext(ctx)

	results.Num = num
	results.Durations = make([]time.Duration, 0, num)

	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	t := time.Now()
	for i := 0; i < num; i++ {
		g.Go(func() error {
			r := bytes.NewReader(data)

			t := time.Now()
			s, err := openai.PostTranscriptionsText(ctx, logCtx, r, "mp3", "")
			if err != nil {
				logCtx.Error("unable get transcription", "error", err)
				return err
			}
			d := time.Since(t)

			if s == "" {
				logCtx.Error("invalid response", "error", err)
				return errors.New("invalid response")
			}

			results.Lock()
			results.Durations = append(results.Durations, d)
			results.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		if !errors.Is(err, context.Canceled) {
			logCtx.Error("unable to process all calls to whisper", "error", err)
		}
	}

	results.Total = time.Since(t)
	results.TotalS = results.Total.String()

	printResults(&results)
	return nil
}
