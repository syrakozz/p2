// Package anthropic integrates with the Anthropic API
package anthropic

import (
	"disruptive/config"
	"disruptive/lib/common"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	baseURL          = "https://api.anthropic.com/v1"
	completeEndpoint = "/complete"
)

// Resty is the shared Resty client for the anthropic package.
var Resty *resty.Client

func init() {
	Resty = resty.New().
		SetBaseURL(baseURL).
		SetLogger(common.LogDiscard).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return r.StatusCode() == http.StatusTooManyRequests
		}).
		SetHeader("x-api-key", config.VARS.AnthropicKey).
		SetHeader("User-Agent", config.VARS.UserAgent).
		SetHeader("Content-Type", "application/json").
		SetTimeout(time.Minute)

	Resty.GetClient().Transport = &http.Transport{
		MaxIdleConnsPerHost: 100,
	}
}
