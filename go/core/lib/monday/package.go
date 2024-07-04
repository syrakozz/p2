// Package monday integrates with the monday.com
package monday

import (
	"disruptive/config"
	"disruptive/lib/common"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	baseURL = "https://api.monday.com/v2"
)

var (
	// Resty is the shared Resty client for the monday package.
	Resty *resty.Client
)

func init() {
	Resty = resty.New().
		SetBaseURL(baseURL).
		SetLogger(common.LogDiscard).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return r.StatusCode() == http.StatusTooManyRequests
		}).
		SetAuthToken(config.VARS.MondayKey).
		SetHeader("User-Agent", config.VARS.UserAgent).
		SetHeader("Content-Type", "application/json").
		SetTimeout(time.Minute)

	Resty.GetClient().Transport = &http.Transport{
		MaxIdleConnsPerHost: 100,
	}
}
