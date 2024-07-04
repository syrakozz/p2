package pinecone

import (
	"context"
	"log/slog"
	"net/http"
)

// UpdateRequest contains fields for a single vector.
type UpdateRequest struct {
	ID          string            `json:"id"`
	Namespace   string            `json:"namespace"`
	SetMetadata map[string]string `json:"setMetadata,omitempty"`
	Values      []float64         `json:"values,omitempty"`
}

// Update modifies an existing vector.
func Update(ctx context.Context, logCtx *slog.Logger, request UpdateRequest) error {
	logCtx = logCtx.With("fid", "pinecone.Update")

	res, err := Resty.R().
		SetContext(ctx).
		SetBody(request).
		Post(updateStarterEndpoint)

	if err != nil {
		return err
	}

	if res.StatusCode() != http.StatusOK {
		errP := res.Error().(*ErrPinecone)
		if errP.Code == 0 {
			errP.Message = res.String()
		}

		logCtx.Error("update endpoint failed", "error", errP.Message)
		return errP
	}

	return nil
}
