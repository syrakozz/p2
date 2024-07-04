package pinecone

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"

	"disruptive/lib/common"
)

// DeleteRequest contains which vectors to delete.
type DeleteRequest struct {
	Namespace string   `json:"namespace"`
	IDs       []string `json:"ids,omitempty"`
	DeleteAll bool     `json:"deleteAll,omitempty"`
	Filter    Metadata `json:"filter,omitempty"`
}

// Delete deletes a vectors.
func Delete(ctx context.Context, logCtx *slog.Logger, request DeleteRequest) error {
	logCtx = logCtx.With("fid", "pinecone.Delete")

	if request.Namespace == "" {
		logCtx.Error("missing namespace")
		return common.ErrBadRequest{Msg: "missing namespace"}
	}

	values := url.Values{}
	values.Set("namespace", request.Namespace)

	if request.IDs != nil {
		values["ids"] = request.IDs
	}

	if request.DeleteAll {
		values.Set("deleteAll", "true")
	}

	res, err := Resty.R().
		SetContext(ctx).
		SetQueryParamsFromValues(values).
		Delete(deleteStarterEndpoint)

	if err != nil {
		return err
	}

	if res.StatusCode() != http.StatusOK {
		errP := res.Error().(*ErrPinecone)
		if errP.Code == 0 {
			errP.Message = res.String()
		}

		logCtx.Error("delete endpoint failed", "error", errP.Message)
		return errP
	}

	return nil
}
