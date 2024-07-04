// Package moderate ...
package moderate

import (
	"context"
	"io/ioutil"
	"log/slog"

	"disruptive/lib/common"
	"disruptive/lib/openai"
)

// ClassifyText returns moderation scores for text or a text file.
func ClassifyText(ctx context.Context, text, filename string) error {
	logCtx := slog.With()

	if filename != "" {
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}

		text = string(b)
	}

	res, err := openai.PostModeration(ctx, logCtx, text)
	if err != nil {
		return err
	}

	common.P(res)
	return nil
}
