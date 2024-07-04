package monday

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"disruptive/lib/common"
)

// GetPing pings monday.com.
func GetPing(ctx context.Context, logCtx *slog.Logger) (bool, error) {
	logCtx = logCtx.With("fid", "monday.GetPing")

	query := `{"query": "{account{name}}"}`

	res, err := Resty.R().
		SetContext(ctx).
		SetBody(query).
		Post("")

	if err != nil {
		logCtx.Error("monday endpoint failed", "error", err)
		return false, err
	}

	if res.StatusCode() == http.StatusUnauthorized {
		logCtx.Error("monday unauthorized", "error", err)
		return false, common.ErrUnauthorized
	}

	if res.StatusCode() != http.StatusOK {
		logCtx.Error("monday endpoint failed", "error", err)
		return false, errors.New(res.Status())
	}

	return true, nil
}
