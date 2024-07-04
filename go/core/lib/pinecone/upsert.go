package pinecone

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

// UpsertRequest contain fields for a slice of vectors.
type UpsertRequest struct {
	Namespace string   `json:"namespace,omitempty"`
	Vectors   []Vector `json:"vectors"`
}

// Upsert adds or updates vectors.
func Upsert(ctx context.Context, logCtx *slog.Logger, request UpsertRequest) ([]string, error) {
	logCtx = logCtx.With("fid", "pinecone.Upsert")

	ids := make([]string, len(request.Vectors))

	for i := 0; i < len(request.Vectors); i++ {
		if request.Vectors[i].ID == "" {
			request.Vectors[i].ID = uuid.New().String()
		}
		ids[i] = request.Vectors[i].ID
	}

	res, err := Resty.R().
		SetContext(ctx).
		SetBody(request).
		Post(upsertStarterEndpoint)

	if err != nil {
		return nil, err
	}

	if res.StatusCode() != http.StatusOK {
		errP := res.Error().(*ErrPinecone)
		if errP.Code == 0 {
			errP.Message = res.String()
		}

		logCtx.Error("stats endpoint failed", "error", errP.Message)
		return nil, errP
	}

	return ids, nil
}
