package pinecone

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
)

// FetchRequest contains fields for fetching vectors.
type FetchRequest struct {
	Namespace string   `json:"namespace"`
	IDs       []string `json:"ids"`
}

// FetchResponse contains fields for fetched vectors.
type FetchResponse struct {
	Namespace string            `json:"namespace"`
	Vectors   map[string]Vector `json:"vectors"`
}

// Fetch retrieves vectors.
func Fetch(ctx context.Context, logCtx *slog.Logger, request FetchRequest) (FetchResponse, error) {
	logCtx = logCtx.With("fid", "pinecone.Fetch")

	res, err := Resty.R().
		SetContext(ctx).
		SetQueryParamsFromValues(url.Values{"namespace": []string{request.Namespace}, "ids": request.IDs}).
		SetResult(&FetchResponse{}).
		Get(fetchStarterEndpoint)

	if err != nil {
		return FetchResponse{}, err
	}

	if res.StatusCode() != http.StatusOK {
		errP := res.Error().(*ErrPinecone)
		if errP.Code == 0 {
			errP.Message = res.String()
		}

		logCtx.Error("fetch endpoint failed", "error", errP.Message)
		return FetchResponse{}, errP
	}

	return *res.Result().(*FetchResponse), nil
}
