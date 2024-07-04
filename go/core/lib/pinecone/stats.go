package pinecone

import (
	"context"
	"log/slog"
	"net/http"
)

// StatsResponse contains index and namespace stats.
type StatsResponse struct {
	Namespaces map[string]struct {
		VectorCount int `json:"vectorCount"`
	} `json:"namespaces"`
	IndexFullness    float64 `json:"indexFullness"`
	TotalVectorCount int     `json:"totalVectorCount"`
}

// Stats returns the number of documents.
func Stats(ctx context.Context, logCtx *slog.Logger, metadata Metadata) (StatsResponse, error) {
	logCtx = logCtx.With("fid", "pinecone.Stats")

	res, err := Resty.R().
		SetContext(ctx).
		SetBody(map[string]Metadata{"filter": metadata}).
		SetResult(&StatsResponse{}).
		Post(statsStarterEndpoint)

	if err != nil {
		return StatsResponse{}, err
	}

	if res.StatusCode() != http.StatusOK {
		errP := res.Error().(*ErrPinecone)
		if errP.Code == 0 {
			errP.Message = res.String()
		}

		logCtx.Error("stats endpoint failed", "error", errP.Message)
		return StatsResponse{}, errP
	}

	return *res.Result().(*StatsResponse), nil
}
