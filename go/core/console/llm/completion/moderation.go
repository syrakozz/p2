package completion

import (
	"context"
	"errors"
	"log/slog"

	"disruptive/lib/openai"
)

func processModeration(ctx context.Context, logCtx *slog.Logger, query string) error {
	m, err := openai.PostModeration(ctx, logCtx, query)
	if err != nil {
		logCtx.Error("unable to get moderation", "error", err)
		return err
	}

	if len(m.Results) != 1 {
		logCtx.Error("invalid moderation results")
		return errors.New("invalid moderation results")
	}

	modResults := make([]any, 0, len(m.Results[0].Categories))
	for k, v := range m.Results[0].Categories {
		if v {
			modResults = append(modResults, k, v)
		}
	}

	if len(modResults) > 0 {
		logCtx.Error("moderation", modResults...)
		return errors.New("moderation failed")
	}

	return nil
}
