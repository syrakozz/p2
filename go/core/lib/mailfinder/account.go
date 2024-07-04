package mailfinder

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
)

// Account contains account data
type Account struct {
	CreditsLeft            int `json:"credits_left"`
	CreditsTotal           int `json:"credits_total"`
	CreditsUsedNotVerified int `json:"credits_used_not_verified"`
	CreditsUsedVerified    int `json:"credits_used_verified"`
}

// GetAccount returns account credit information
func GetAccount(ctx context.Context, logCtx *slog.Logger) (Account, error) {
	logCtx = logCtx.With("fid", "mailfinder.GetAccount")

	res, err := Resty.R().
		SetContext(ctx).
		SetResult(&Account{}).
		Get(anymailfinderAccountEndpoint)

	if err != nil {
		logCtx.Error("anymailfinder account endpoint failed", "error", err)
		return Account{}, err
	}

	if res.StatusCode() != http.StatusOK {
		logCtx.Error("anymailfinder endpoint failed", "status", res.Status())
		return Account{}, errors.New(res.Status())
	}

	return *res.Result().(*Account), nil
}
