package stress

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"time"

	"golang.org/x/sync/errgroup"

	"disruptive/lib/elevenlabs"
)

// Elevenlabs stress tests the Elevenlabs TTS API.
func Elevenlabs(ctx context.Context, text string, num int) error {
	logCtx := slog.With("num", num)

	g, ctx := errgroup.WithContext(ctx)

	results.Num = num
	results.Durations = make([]time.Duration, 0, num)

	req := elevenlabs.Request{
		Text:              text,
		Voice:             "rachel",
		Stability:         75,
		SimilarityBoost:   75,
		StyleExaggeration: 0,
	}

	t := time.Now()
	for i := 0; i < num; i++ {
		g.Go(func() error {
			t := time.Now()

			r, err := elevenlabs.TTSStream(ctx, logCtx, req)
			if err != nil {
				logCtx.Error("unable to get TTS stream", "error", err)
				return err
			}

			n, err := io.Copy(io.Discard, r)
			if err != nil {
				logCtx.Error("unable to copy response", "error", err)
				return err
			}

			if n == 0 {
				logCtx.Error("empty response")
				return err
			}

			d := time.Since(t)

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
