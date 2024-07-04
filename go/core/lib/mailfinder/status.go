package mailfinder

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
)

type status struct {
	Healthy bool `json:"healthy"`
}

// GetStatus returns the status of the anymailfinder service
func GetStatus(ctx context.Context, logCtx *slog.Logger) (bool, error) {
	logCtx = logCtx.With("fid", "mailfinder.GetStatus")

	res, err := Resty.R().
		SetContext(ctx).
		SetResult(&status{}).
		Get(anymailfinderStatusEndpoint)

	if err != nil {
		logCtx.Error("anymailfinder status endpoint failed", "error", err)
		return false, err
	}

	if res.StatusCode() != http.StatusOK {
		logCtx.Error("anymailfinder status endpoint failed", "status", res.Status())
		return false, errors.New(res.Status())
	}

	return res.Result().(*status).Healthy, nil
}
