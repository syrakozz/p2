package pinecone

import (
	"context"
	"log/slog"
	"net/http"
)

// QueryRequest contains information about the query vector.
// Each query request can contain only one of vector or id.
type QueryRequest struct {
	ID              string    `json:"id"`
	Namespace       string    `json:"namespace"`
	TopK            int       `json:"topK"`
	Filter          Metadata  `json:"filter"`
	Vector          []float64 `json:"vector"`
	IncludeValues   bool      `json:"includeValues"`
	IncludeMetadata bool      `json:"includeMetadata"`
}

// QueryResponse contains a single reponse document.
type QueryResponse struct {
	ID       string    `json:"id"`
	Score    float64   `json:"score"`
	Values   []float64 `json:"values,omitempty"`
	Metadata Metadata  `json:"metadata,omitempty"`
}

// QueryResponses contains that the TopK nearest results.
type QueryResponses struct {
	Namespace string          `json:"namespace"`
	Matches   []QueryResponse `json:"matches"`
}

// Query retrieves vectors.
func Query(ctx context.Context, logCtx *slog.Logger, query QueryRequest) ([]QueryResponse, error) {
	logCtx = logCtx.With("fid", "pinecone.Query")

	res, err := Resty.R().
		SetContext(ctx).
		SetBody(query).
		SetResult(&QueryResponses{}).
		Post(queryStarterEndpoint)

	if err != nil {
		return nil, err
	}

	if res.StatusCode() != http.StatusOK {
		errP := res.Error().(*ErrPinecone)
		if errP.Code == 0 {
			errP.Message = res.String()
		}

		logCtx.Error("query endpoint failed", "error", errP.Message)
		return nil, errP
	}

	return res.Result().(*QueryResponses).Matches, nil
}
