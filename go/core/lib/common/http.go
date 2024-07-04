// Package common includes shared functionality.
package common

import (
	"context"
	"net/http"

	"github.com/go-resty/resty/v2"
	"google.golang.org/api/idtoken"

	"disruptive/config"
)

type logDiscard struct{}

func (l logDiscard) Debugf(_ string, _ ...any) {}

func (l logDiscard) Warnf(_ string, _ ...any) {}

func (l logDiscard) Errorf(_ string, _ ...any) {}

// LogDiscard ...
var LogDiscard logDiscard

// NewIAPResty returns a resty client with IAP authentication.
func NewIAPResty(ctx context.Context, audience string) (*resty.Client, error) {
	httpClient, err := idtoken.NewClient(ctx, audience)
	if err != nil {
		return nil, err
	}

	return resty.NewWithClient(httpClient).
			SetLogger(LogDiscard).
			AddRetryCondition(func(r *resty.Response, err error) bool {
				return r.StatusCode() == http.StatusTooManyRequests
			}).
			SetHeader("User-Agent", config.VARS.UserAgent).
			SetHeader("Content-Type", "application/json"),
		nil
}

// NewHTTPResty returns a resty created form an existing http.Client.
func NewHTTPResty(client *http.Client) *resty.Client {
	return resty.NewWithClient(client).
		SetLogger(LogDiscard).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return r.StatusCode() == http.StatusTooManyRequests
		}).
		SetHeader("User-Agent", config.VARS.UserAgent).
		SetHeader("Content-Type", "application/json")
}
