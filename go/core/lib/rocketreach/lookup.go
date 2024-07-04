package rocketreach

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
)

// LookupRequest contains API request fields.
type LookupRequest struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	CurrentEmployer string `json:"current_employer"`
	Title           string `json:"title"`
	LinkedInURL     string `json:"linkedin_url"`
	Email           string `json:"email"`
}

// LookupResponse contains API response fields.
type LookupResponse map[string]any

func renderFromLookupRequest(request LookupRequest) url.Values {
	v := url.Values{}

	if request.ID != "" {
		v.Set("id", request.ID)
	}

	if request.Name != "" {
		v.Set("name", request.Name)
	}

	if request.CurrentEmployer != "" {
		v.Set("current_employer", request.CurrentEmployer)
	}

	if request.Title != "" {
		v.Set("title", request.Title)
	}

	if request.LinkedInURL != "" {
		v.Set("linkedin_url", request.LinkedInURL)
	}

	if request.Email != "" {
		v.Set("email", request.Email)
	}

	v.Set("lookup_type", "enrich")

	return v
}

// GetLookup does a person and company lookup.
// https://rocketreach.co/api/docs/#operation%2Fprofile-company_lookup_read
func GetLookup(ctx context.Context, logCtx *slog.Logger, request LookupRequest) (LookupResponse, error) {
	logCtx = logCtx.With("fid", "rocketreach.GetLookup")

	res, err := Resty.R().
		SetContext(ctx).
		SetQueryParamsFromValues(renderFromLookupRequest(request)).
		SetResult(&LookupResponse{}).
		Get(lookupEndpoint)

	if err != nil {
		logCtx.Error("rocketreach lookup endpoint failed", "error", err)
		return nil, err
	}

	if res.StatusCode() != http.StatusOK {
		logCtx.Error("rocketreach lookup endpoint failed", "status", res.Status())
		return nil, errors.New(res.Status())
	}

	return *res.Result().(*LookupResponse), nil
}
