package rocketreach

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
)

// SearchRequest contains API request fields.
type SearchRequest map[string]any

// SearchResponse contains API response fields.
type SearchResponse map[string]any

func renderFromSearchRequest(req SearchRequest) map[string]any {
	sr := SearchRequest{}

	start, ok := req["start"]
	if ok {
		delete(req, "start")
	} else {
		start = 1
	}

	// Change single element into a slice of elements
	for k, v := range req {
		switch t := v.(type) {
		case []any:
			sr[k] = t
		case any:
			sr[k] = []any{t}
		default:
			continue
		}
	}

	return map[string]any{"start": start, "query": sr}
}

// PostSearch searches for people.
// https://rocketreach.co/api/docs/#operation/person_search_create
func PostSearch(ctx context.Context, logCtx *slog.Logger, request SearchRequest) (SearchResponse, error) {
	logCtx = logCtx.With("fid", "rocketreach.PostSearch")

	_request := renderFromSearchRequest(request)

	res, err := Resty.R().
		SetContext(ctx).
		SetBody(_request).
		SetResult(&SearchResponse{}).
		Post(searchEndpoint)

	if err != nil {
		logCtx.Error("rocketreach search endpoint failed", "error", err)
		return nil, err
	}

	if res.StatusCode() != http.StatusCreated {
		logCtx.Error("rocketreach search endpoint failed", "status", res.Status())
		return nil, errors.New(res.Status())
	}

	return *res.Result().(*SearchResponse), nil
}
