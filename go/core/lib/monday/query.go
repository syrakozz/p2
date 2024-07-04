package monday

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"disruptive/lib/common"
)

// PostQuery proxies a GraphQL query to monday.com and returns and JSON response.
func PostQuery(ctx context.Context, logCtx *slog.Logger, query []byte) ([]byte, error) {
	logCtx = logCtx.With("fid", "monday.PostQuery")

	res, err := Resty.R().
		SetContext(ctx).
		SetBody(query).
		Post("")

	if err != nil {
		logCtx.Error("monday endpoint failed", "error", err)
		return nil, err
	}

	if res.StatusCode() == http.StatusForbidden {
		logCtx.Error("monday forbidden", "status", res.Status())
		return nil, common.ErrForbidden{Msg: res.Status()}
	}

	if res.StatusCode() != http.StatusOK {
		logCtx.Error("monday endpoint failed", "status", res.Status())
		return nil, errors.New(res.Status())
	}

	return res.Body(), nil
}
