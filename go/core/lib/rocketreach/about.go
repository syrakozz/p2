package rocketreach

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
)

// About contains about fields.
// Types defined as any can be an int or string "inf"
type About struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	State     string `json:"state"`
	Plan      struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		LookupLimit any    `json:"lookup_limit"`
		ExportLimit int    `json:"export_limit"`
	} `json:"plan"`
	APIKey               string `json:"api_key"`
	APIKeyDomain         string `json:"api_key_domain"`
	LookupCreditBalance  any    `json:"lookup_credit_balance"`
	ExportCreditBalance  int    `json:"export_credit_balance"`
	LifetimeCreditsSpent int    `json:"lifetime_credits_spent"`
	LifetimeAPINumCalls  int    `json:"lifetime_api_num_calls"`
	DailyAPINumCalls     int    `json:"daily_api_num_calls"`
	DailyAPILimit        int    `json:"daily_api_limit"`
}

// GetAbout returns account information.
// https://rocketreach.co/api/docs/#operation%2Faccount_read
func GetAbout(ctx context.Context, logCtx *slog.Logger) (About, error) {
	logCtx = logCtx.With("fid", "rocketreach.GetAbout")

	res, err := Resty.R().
		SetContext(ctx).
		SetResult(&About{}).
		Get(aboutEndpoint)

	if err != nil {
		logCtx.Error("rocketreach about endpoint failed", "error", err)
		return About{}, err
	}

	if res.StatusCode() != http.StatusOK {
		logCtx.Error("rocketreach about endpoint failed", "status", res.Status())
		return About{}, errors.New(res.Status())
	}

	return *res.Result().(*About), nil
}
